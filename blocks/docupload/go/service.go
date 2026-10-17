package docupload

import (
	"bytes"
	"context"
	"encoding/hex"
	"path"
	"strings"
	"time"
	"unicode"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// CompleteFunc runs after the last block of a session is stored.
type CompleteFunc func(ctx context.Context, accountID, fileID string)

// Options configures a Service.
type Options struct {
	Config     Config
	Logger     *zap.Logger
	OnComplete CompleteFunc
	Now        func() time.Time
}

// Service is the upload domain API.
type Service struct {
	store      *Store
	cfg        Config
	logger     *zap.Logger
	onComplete CompleteFunc
	now        func() time.Time
}

// New builds a service backed by Redis. rdb may be nil only in tests that never call it.
func New(rdb *redis.Client, opts Options) *Service {
	cfg := opts.Config.normalized()
	logger := opts.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	return &Service{
		store:      NewStore(rdb),
		cfg:        cfg,
		logger:     logger,
		onComplete: opts.OnComplete,
		now:        now,
	}
}

// StartUpload opens a session or refreshes one with the same compressed fingerprint.
func (s *Service) StartUpload(ctx context.Context, accountID string, in StartRequest) (StartResult, error) {
	if accountID == "" {
		return StartResult{}, ErrUnauthenticated
	}
	fileID, err := normalizeMD5(in.FileID)
	if err != nil {
		return StartResult{}, err
	}
	compressedMD5, err := normalizeMD5(in.CompressedMD5)
	if err != nil {
		return StartResult{}, err
	}
	if in.OriginalSize <= 0 {
		return StartResult{}, ErrEmptyFile
	}
	if in.OriginalSize > s.cfg.MaxOriginalBytes || in.CompressedSize > s.cfg.MaxCompressedBytes {
		return StartResult{}, ErrTooLarge
	}
	if in.CompressedSize <= 0 {
		return StartResult{}, ErrInvalidArgument.With("compressed_size", "must be positive")
	}
	name, err := cleanName(in.FileName)
	if err != nil {
		return StartResult{}, err
	}
	expires, ttl, err := s.expiry(in.ExpiresUnixMs)
	if err != nil {
		return StartResult{}, err
	}

	existing, err := s.store.Get(ctx, accountID, fileID)
	if err != nil && err != ErrNotFound && err != ErrExpired {
		return StartResult{}, wrapErr(ErrInternal, "get", err)
	}
	if err == nil &&
		existing.CompressedMD5 == compressedMD5 &&
		existing.CompressedSize == in.CompressedSize &&
		existing.BlockSize == s.cfg.BlockSize {
		if err := s.store.Touch(ctx, existing, expires, ttl); err != nil {
			return StartResult{}, wrapErr(ErrInternal, "touch", err)
		}
		pending, err := s.pending(ctx, accountID, fileID, existing.BlockCount)
		if err != nil {
			return StartResult{}, err
		}
		return StartResult{
			BlockSize:     int32(existing.BlockSize),
			Pending:       pending,
			MaxThreads:    int32(existing.MaxThreads),
			ExpiresUnixMs: expires,
		}, nil
	}

	count := blockCountOf(in.CompressedSize, s.cfg.BlockSize)
	session := Session{
		AccountID:      accountID,
		FileID:         fileID,
		FileName:       name,
		FileUnixMs:     in.FileUnixMs,
		OriginalSize:   in.OriginalSize,
		CompressedMD5:  compressedMD5,
		CompressedSize: in.CompressedSize,
		BlockSize:      s.cfg.BlockSize,
		BlockCount:     count,
		MaxThreads:     s.cfg.MaxThreads,
		ExpiresUnixMs:  expires,
		ContentType:    cleanContentType(in.ContentType),
		Source:         cleanSource(in.Source),
	}
	if err := s.store.SaveNew(ctx, session, ttl); err != nil {
		return StartResult{}, wrapErr(ErrInternal, "save", err)
	}
	return StartResult{
		BlockSize:     int32(session.BlockSize),
		Pending:       allIndexes(count),
		MaxThreads:    int32(session.MaxThreads),
		ExpiresUnixMs: expires,
	}, nil
}

// UploadBlock checks the slice and stores it. A digest or length mismatch returns Success false.
func (s *Service) UploadBlock(ctx context.Context, accountID string, in BlockRequest) (BlockResult, error) {
	if accountID == "" {
		return BlockResult{}, ErrUnauthenticated
	}
	fileID, err := normalizeMD5(in.FileID)
	if err != nil {
		return BlockResult{}, err
	}
	blockMD5, err := normalizeMD5(in.BlockMD5)
	if err != nil {
		return BlockResult{}, err
	}
	session, err := s.store.Get(ctx, accountID, fileID)
	if err != nil {
		if err == ErrNotFound || err == ErrExpired {
			return BlockResult{}, err
		}
		return BlockResult{}, wrapErr(ErrInternal, "get", err)
	}
	if md5Hex(in.Payload) != blockMD5 {
		pending, err := s.pending(ctx, accountID, fileID, session.BlockCount)
		if err != nil {
			return BlockResult{}, err
		}
		return BlockResult{Success: false, Pending: pending}, nil
	}

	ttl := time.Until(time.UnixMilli(session.ExpiresUnixMs))
	status, added, received, err := s.store.PutBlock(ctx, accountID, fileID, in.Index, in.Payload, s.now(), ttl)
	if err != nil {
		return BlockResult{}, wrapErr(ErrInternal, "put", err)
	}
	switch status {
	case "missing":
		return BlockResult{}, ErrNotFound
	case "expired":
		return BlockResult{}, ErrExpired
	case "conflict":
		return BlockResult{}, ErrConflict
	case "index", "length":
		pending, err := s.pending(ctx, accountID, fileID, session.BlockCount)
		if err != nil {
			return BlockResult{}, err
		}
		return BlockResult{Success: false, Pending: pending}, nil
	case "ok":
		pending, err := s.pending(ctx, accountID, fileID, session.BlockCount)
		if err != nil {
			return BlockResult{}, err
		}
		if added == 1 && received == session.BlockCount {
			s.fireComplete(accountID, fileID)
		}
		return BlockResult{Success: true, Pending: pending}, nil
	default:
		return BlockResult{}, ErrInternal.With("status", status)
	}
}

// CancelUpload removes the session for this account and file id.
func (s *Service) CancelUpload(ctx context.Context, accountID, fileID string) error {
	if accountID == "" {
		return ErrUnauthenticated
	}
	id, err := normalizeMD5(fileID)
	if err != nil {
		return err
	}
	if _, err := s.store.Get(ctx, accountID, id); err != nil && err != ErrExpired {
		return err
	}
	if err := s.store.Delete(ctx, accountID, id); err != nil {
		return wrapErr(ErrInternal, "delete", err)
	}
	return nil
}

// GetCompleteDoc rebuilds the original file when every block is present.
func (s *Service) GetCompleteDoc(ctx context.Context, accountID, fileID string) (CompleteDoc, error) {
	if accountID == "" {
		return CompleteDoc{}, ErrUnauthenticated
	}
	id, err := normalizeMD5(fileID)
	if err != nil {
		return CompleteDoc{}, err
	}
	session, err := s.store.Get(ctx, accountID, id)
	if err != nil {
		return CompleteDoc{}, err
	}
	got, err := s.store.Received(ctx, accountID, id)
	if err != nil {
		return CompleteDoc{}, wrapErr(ErrInternal, "received", err)
	}
	if len(got) != session.BlockCount {
		return CompleteDoc{}, ErrIncomplete
	}
	var compressed bytes.Buffer
	compressed.Grow(int(session.CompressedSize))
	for i := 0; i < session.BlockCount; i++ {
		part, err := s.store.Block(ctx, accountID, id, int32(i))
		if err != nil {
			return CompleteDoc{}, ErrIncomplete
		}
		compressed.Write(part)
	}
	raw := compressed.Bytes()
	if int64(len(raw)) != session.CompressedSize || md5Hex(raw) != session.CompressedMD5 {
		return CompleteDoc{}, ErrChecksum
	}
	plain, err := gunzipLimited(raw, s.cfg.MaxOriginalBytes)
	if err != nil {
		if err == ErrTooLarge {
			return CompleteDoc{}, err
		}
		return CompleteDoc{}, ErrChecksum
	}
	if md5Hex(plain) != session.FileID || int64(len(plain)) != session.OriginalSize {
		return CompleteDoc{}, ErrChecksum
	}
	return CompleteDoc{
		FileID:         session.FileID,
		FileName:       session.FileName,
		FileUnixMs:     session.FileUnixMs,
		OriginalSize:   session.OriginalSize,
		CompressedMD5:  session.CompressedMD5,
		CompressedSize: session.CompressedSize,
		ExpiresUnixMs:  session.ExpiresUnixMs,
		ContentType:    session.ContentType,
		Source:         session.Source,
		AccountID:      session.AccountID,
		Bytes:          plain,
	}, nil
}

// Sweep drops stale sessions. The background worker calls this about once a minute.
func (s *Service) Sweep(ctx context.Context, now time.Time) (int, error) {
	n, err := s.store.Sweep(ctx, now)
	if err != nil {
		return n, wrapErr(ErrInternal, "sweep", err)
	}
	return n, nil
}

func (s *Service) fireComplete(accountID, fileID string) {
	if s.onComplete == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		s.onComplete(ctx, accountID, fileID)
	}()
}

func (s *Service) pending(ctx context.Context, account, fileID string, count int) ([]int32, error) {
	got, err := s.store.Received(ctx, account, fileID)
	if err != nil {
		return nil, wrapErr(ErrInternal, "pending", err)
	}
	return pendingIndexes(count, got), nil
}

func (s *Service) expiry(requestedUnixMs int64) (int64, time.Duration, error) {
	now := s.now()
	ttl := s.cfg.DefaultTTL
	if requestedUnixMs > 0 {
		requested := time.UnixMilli(requestedUnixMs)
		if !requested.After(now) {
			return 0, 0, ErrExpired
		}
		ttl = requested.Sub(now)
		if ttl > s.cfg.MaxTTL {
			ttl = s.cfg.MaxTTL
		}
	}
	expires := now.Add(ttl).UnixMilli()
	return expires, ttl, nil
}

func normalizeMD5(raw string) (string, error) {
	sum := strings.ToLower(strings.TrimSpace(raw))
	if len(sum) != 32 {
		return "", ErrInvalidArgument.With("md5", "must be 32 hex characters")
	}
	if _, err := hex.DecodeString(sum); err != nil {
		return "", ErrInvalidArgument.With("md5", "must be hex")
	}
	return sum, nil
}

func cleanName(name string) (string, error) {
	name = strings.ReplaceAll(name, "\\", "/")
	name = path.Base(name)
	name = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 || unicode.IsControl(r) {
			return -1
		}
		return r
	}, strings.TrimSpace(name))
	if name == "" || name == "." || name == ".." {
		return "", ErrInvalidArgument.With("file_name", "required")
	}
	if len(name) > 255 {
		name = name[:255]
	}
	return name, nil
}

func cleanContentType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "application/octet-stream"
	}
	if len(value) > 128 {
		return value[:128]
	}
	return value
}

func cleanSource(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "", "drop":
		return "drop"
	case "camera", "gallery":
		return value
	default:
		if len(value) > 32 {
			return value[:32]
		}
		return value
	}
}
