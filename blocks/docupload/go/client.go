package docupload

import (
	"context"
	"net/http"
	"sync"
	"time"

	"connectrpc.com/connect"

	docuploadv1 "github.com/emersonary/appkit/docupload/gen/docupload/v1"
	"github.com/emersonary/appkit/docupload/gen/docupload/v1/docuploadv1connect"
)

const (
	defaultRetries = 5
	maxBackoff     = 8 * time.Second
)

// UploadFile is a local document the client compresses and sends.
type UploadFile struct {
	Name        string
	Bytes       []byte
	ContentType string
	Source      string
	FileTime    time.Time
	ExpiresAt   time.Time
}

// Prepared is the hashed, gzip-compressed form of an upload.
type Prepared struct {
	File          UploadFile
	FileID        string
	Compressed    []byte
	CompressedMD5 string
	FileUnixMs    int64
	ExpiresUnixMs int64
}

// Client calls the Connect upload service.
type Client struct {
	baseURL    string
	httpClient connect.HTTPClient
	accountID  string
	retries    int
	backoff    time.Duration
}

// NewClient builds a consumer of DocUploadService.
// accountID is sent as the bearer token; the host resolver maps that token to an account.
func NewClient(baseURL string, httpClient connect.HTTPClient, accountID string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		accountID:  accountID,
		retries:    defaultRetries,
		backoff:    200 * time.Millisecond,
	}
}

// WithRetries overrides the per-block attempt count. Values below 1 become 1.
func (c *Client) WithRetries(n int) *Client {
	if n < 1 {
		n = 1
	}
	c.retries = n
	return c
}

// WithBackoff sets the base delay between attempts. Zero disables the wait.
func (c *Client) WithBackoff(d time.Duration) *Client {
	if d < 0 {
		d = 0
	}
	c.backoff = d
	return c
}

// Prepare hashes and gzip-compresses the file. It does not call the server.
func Prepare(file UploadFile) (Prepared, error) {
	compressed, err := Gzip(file.Bytes)
	if err != nil {
		return Prepared{}, err
	}
	fileUnix := file.FileTime.UnixMilli()
	if file.FileTime.IsZero() {
		fileUnix = time.Now().UnixMilli()
	}
	var expires int64
	if !file.ExpiresAt.IsZero() {
		expires = file.ExpiresAt.UnixMilli()
	}
	return Prepared{
		File:          file,
		FileID:        md5Hex(file.Bytes),
		Compressed:    compressed,
		CompressedMD5: md5Hex(compressed),
		FileUnixMs:    fileUnix,
		ExpiresUnixMs: expires,
	}, nil
}

// Start sends startUpload for an already prepared body.
func (c *Client) Start(ctx context.Context, prepped Prepared) (StartResult, error) {
	req := connect.NewRequest(&docuploadv1.StartUploadRequest{
		FileId:         prepped.FileID,
		FileName:       prepped.File.Name,
		FileUnixMs:     prepped.FileUnixMs,
		OriginalSize:   int64(len(prepped.File.Bytes)),
		CompressedMd5:  prepped.CompressedMD5,
		CompressedSize: int64(len(prepped.Compressed)),
		ExpiresUnixMs:  prepped.ExpiresUnixMs,
		ContentType:    prepped.File.ContentType,
		Source:         prepped.File.Source,
	})
	c.authorize(req.Header())
	resp, err := c.rpc().StartUpload(ctx, req)
	if err != nil {
		return StartResult{}, err
	}
	msg := resp.Msg
	return StartResult{
		BlockSize:     msg.GetBlockSize(),
		Pending:       append([]int32(nil), msg.GetPendingBlocks()...),
		MaxThreads:    msg.GetMaxThreads(),
		ExpiresUnixMs: msg.GetExpiresUnixMs(),
	}, nil
}

// PutBlock sends one slice. sum is the block MD5; an empty sum is computed from payload.
func (c *Client) PutBlock(ctx context.Context, fileID string, index int32, payload []byte, sum string) (BlockResult, error) {
	if sum == "" {
		sum = md5Hex(payload)
	}
	req := connect.NewRequest(&docuploadv1.UploadBlockRequest{
		FileId:   fileID,
		Index:    index,
		Payload:  payload,
		BlockMd5: sum,
	})
	c.authorize(req.Header())
	resp, err := c.rpc().UploadBlock(ctx, req)
	if err != nil {
		return BlockResult{}, err
	}
	return BlockResult{
		Success: resp.Msg.GetSuccess(),
		Pending: append([]int32(nil), resp.Msg.GetPendingBlocks()...),
	}, nil
}

// PutBlockRetry repeats PutBlock until it succeeds or the attempt budget is spent.
func (c *Client) PutBlockRetry(ctx context.Context, fileID string, index int32, payload []byte) error {
	var last error
	for attempt := 0; attempt < c.retries; attempt++ {
		if attempt > 0 {
			if err := c.pause(ctx, attempt-1); err != nil {
				return err
			}
		}
		result, err := c.PutBlock(ctx, fileID, index, payload, "")
		if err != nil {
			last = err
			continue
		}
		if !result.Success {
			last = ErrChecksum
			continue
		}
		return nil
	}
	if last == nil {
		last = ErrChecksum
	}
	return last
}

// Cancel deletes the server session.
func (c *Client) Cancel(ctx context.Context, fileID string) error {
	req := connect.NewRequest(&docuploadv1.CancelUploadRequest{FileId: fileID})
	c.authorize(req.Header())
	_, err := c.rpc().CancelUpload(ctx, req)
	return err
}

// Upload compresses the file, starts a session, and sends every pending block.
func (c *Client) Upload(ctx context.Context, file UploadFile) (Prepared, StartResult, error) {
	prepped, err := Prepare(file)
	if err != nil {
		return Prepared{}, StartResult{}, err
	}
	started, err := c.Start(ctx, prepped)
	if err != nil {
		return prepped, StartResult{}, err
	}
	if err := c.SendPending(ctx, prepped, started); err != nil {
		return prepped, started, err
	}
	return prepped, started, nil
}

// SendPending uploads only the indexes in started.Pending.
func (c *Client) SendPending(ctx context.Context, prepped Prepared, started StartResult) error {
	blocks := SplitBlocks(prepped.Compressed, int(started.BlockSize))
	threads := int(started.MaxThreads)
	if threads < 1 {
		threads = 1
	}
	if len(started.Pending) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan int32)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for index := range jobs {
			if ctx.Err() != nil {
				return
			}
			if int(index) < 0 || int(index) >= len(blocks) {
				select {
				case errCh <- ErrInvalidArgument:
				default:
				}
				cancel()
				return
			}
			if err := c.PutBlockRetry(ctx, prepped.FileID, index, blocks[index]); err != nil {
				select {
				case errCh <- err:
				default:
				}
				cancel()
				return
			}
		}
	}
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go worker()
	}
	go func() {
		defer close(jobs)
		for _, index := range started.Pending {
			select {
			case <-ctx.Done():
				return
			case jobs <- index:
			}
		}
	}()
	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
	}
	return ctx.Err()
}

func (c *Client) rpc() docuploadv1connect.DocUploadServiceClient {
	return docuploadv1connect.NewDocUploadServiceClient(c.httpClient, c.baseURL)
}

func (c *Client) authorize(header http.Header) {
	if c.accountID != "" {
		header.Set("Authorization", "Bearer "+c.accountID)
	}
}

func (c *Client) pause(ctx context.Context, attempt int) error {
	delay := c.backoff
	if delay > 0 {
		delay <<= attempt
		if delay > maxBackoff || delay <= 0 {
			delay = maxBackoff
		}
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
