package docupload_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alicebob/miniredis/v2"
	"github.com/emersonary/appkit/apperror"
	"github.com/emersonary/appkit/docupload"
	"github.com/emersonary/appkit/docupload/transport"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type testEnv struct {
	svc       *docupload.Service
	srv       *httptest.Server
	rdb       *redis.Client
	completed atomic.Int32
}

func newTestEnv(t *testing.T, wrap func(http.Handler) http.Handler) *testEnv {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	env := &testEnv{rdb: rdb}
	env.svc = docupload.New(rdb, docupload.Options{
		Logger: zap.NewNop(),
		OnComplete: func(context.Context, string, string) {
			env.completed.Add(1)
		},
	})
	mux := http.NewServeMux()
	transport.New(env.svc, &transport.Mount{
		HTTPMux: mux,
		ResolveAccountID: func(_ context.Context, token string) (string, error) {
			if token == "" {
				return "", docupload.ErrUnauthenticated
			}
			return token, nil
		},
	})
	var handler http.Handler = mux
	if wrap != nil {
		handler = wrap(mux)
	}
	env.srv = httptest.NewServer(handler)
	t.Cleanup(env.srv.Close)
	return env
}

func (e *testEnv) client(account string) *docupload.Client {
	return docupload.NewClient(e.srv.URL, e.srv.Client(), account).WithBackoff(0)
}

func connectCode(err error) connect.Code {
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return connectErr.Code()
	}
	return 0
}

func appCode(err error) string {
	var appErr apperror.Error
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ""
}

func waitComplete(t *testing.T, env *testEnv, want int32) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for env.completed.Load() < want {
		if time.Now().After(deadline) {
			t.Fatalf("onComplete = %d, want %d", env.completed.Load(), want)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func assertDoc(t *testing.T, doc docupload.CompleteDoc, prepped docupload.Prepared) {
	t.Helper()
	if !bytes.Equal(doc.Bytes, prepped.File.Bytes) {
		t.Fatalf("assembled bytes = %d, original = %d", len(doc.Bytes), len(prepped.File.Bytes))
	}
	if doc.FileID != prepped.FileID || doc.FileID != md5Hex(prepped.File.Bytes) {
		t.Fatalf("file id = %s, want %s", doc.FileID, prepped.FileID)
	}
	if doc.CompressedMD5 != prepped.CompressedMD5 || doc.CompressedMD5 != md5Hex(prepped.Compressed) {
		t.Fatalf("compressed md5 = %s, want %s", doc.CompressedMD5, prepped.CompressedMD5)
	}
	if doc.OriginalSize != int64(len(prepped.File.Bytes)) {
		t.Fatalf("original size = %d, want %d", doc.OriginalSize, len(prepped.File.Bytes))
	}
	if doc.CompressedSize != int64(len(prepped.Compressed)) {
		t.Fatalf("compressed size = %d, want %d", doc.CompressedSize, len(prepped.Compressed))
	}
	if doc.FileUnixMs != prepped.FileUnixMs {
		t.Fatalf("file time = %d, want %d", doc.FileUnixMs, prepped.FileUnixMs)
	}
	if doc.FileName != cleanExpectedName(prepped.File.Name) {
		t.Fatalf("file name = %q, want %q", doc.FileName, cleanExpectedName(prepped.File.Name))
	}
}

func md5Hex(b []byte) string {
	sum := md5.Sum(b)
	return hex.EncodeToString(sum[:])
}

func gzipNoCompression(b []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, gzip.NoCompression)
	if err != nil {
		return nil, err
	}
	w.ModTime = time.Unix(0, 0).UTC()
	w.OS = 255
	if _, err := w.Write(b); err != nil {
		_ = w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func cleanExpectedName(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	return strings.TrimSpace(pathBase(name))
}

func pathBase(name string) string {
	if i := strings.LastIndex(name, "/"); i >= 0 {
		return name[i+1:]
	}
	return name
}

func containsIndex(pending []int32, index int32) bool {
	for _, item := range pending {
		if item == index {
			return true
		}
	}
	return false
}

func smallJPEG() []byte {
	return []byte{
		0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00,
		0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00,
		0xFF, 0xD9,
	}
}

func sizedPDF(n int) []byte {
	return framed(n, []byte("%PDF-1.4\n"), []byte("\n%%EOF"), 7)
}

func zeroFilledJPEG(n int) []byte {
	buf := make([]byte, n)
	buf[0], buf[1] = 0xFF, 0xD8
	buf[n-2], buf[n-1] = 0xFF, 0xD9
	return buf
}

func framed(n int, head, tail []byte, seed int64) []byte {
	if n < len(head)+len(tail) {
		panic("frame larger than buffer")
	}
	buf := make([]byte, n)
	rng := rand.New(rand.NewSource(seed))
	_, _ = rng.Read(buf)
	copy(buf, head)
	copy(buf[n-len(tail):], tail)
	return buf
}

func writeFixtures(t *testing.T) (map[string][]byte, string) {
	t.Helper()
	dir := t.TempDir()
	files := map[string][]byte{
		"empty.pdf":    {},
		"photo.jpg":    smallJPEG(),
		"doc-19mb.pdf": sizedPDF(19 << 20),
		"too-big.jpg":  zeroFilledJPEG((20 << 20) + 1),
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	loaded := make(map[string][]byte, len(files))
	for name := range files {
		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		loaded[name] = body
	}
	return loaded, dir
}

func uploadFile(name, contentType, source string, body []byte, when time.Time) docupload.UploadFile {
	return docupload.UploadFile{
		Name:        name,
		Bytes:       body,
		ContentType: contentType,
		Source:      source,
		FileTime:    when,
	}
}

func TestClientFixtures(t *testing.T) {
	files, _ := writeFixtures(t)
	when := time.UnixMilli(1_700_000_000_000).UTC()

	t.Run("empty", func(t *testing.T) {
		env := newTestEnv(t, nil)
		_, _, err := env.client("acct-1").Upload(context.Background(), uploadFile("empty.pdf", "application/pdf", "drop", files["empty.pdf"], when))
		if connectCode(err) != connect.CodeInvalidArgument {
			t.Fatalf("empty upload code = %s, err = %v", connectCode(err), err)
		}
		if n := len(env.mrKeys(t)); n != 0 {
			t.Fatalf("redis keys = %d", n)
		}
	})

	t.Run("jpeg", func(t *testing.T) {
		env := newTestEnv(t, nil)
		client := env.client("acct-1")
		file := uploadFile("photo.jpg", "image/jpeg", "camera", files["photo.jpg"], when)
		prepped, started, err := client.Upload(context.Background(), file)
		if err != nil {
			t.Fatal(err)
		}
		if started.BlockSize != docupload.DefaultBlockSize || started.MaxThreads != docupload.DefaultMaxThreads {
			t.Fatalf("block policy = %d threads %d", started.BlockSize, started.MaxThreads)
		}
		doc, err := env.svc.GetCompleteDoc(context.Background(), "acct-1", prepped.FileID)
		if err != nil {
			t.Fatal(err)
		}
		assertDoc(t, doc, prepped)
		if doc.ContentType != "image/jpeg" || doc.Source != "camera" || doc.AccountID != "acct-1" {
			t.Fatalf("meta = %+v", doc)
		}
		waitComplete(t, env, 1)
	})

	t.Run("pdf19", func(t *testing.T) {
		env := newTestEnv(t, nil)
		body := files["doc-19mb.pdf"]
		var prepped docupload.Prepared
		var err error
		for i := 0; i < 5; i++ {
			prepped, err = docupload.Prepare(uploadFile("doc-19mb.pdf", "application/pdf", "drop", body, when))
			if err != nil {
				t.Fatal(err)
			}
			if len(prepped.Compressed)%docupload.DefaultBlockSize != 0 {
				break
			}
			body[64+i]++
			prepped, err = docupload.Prepare(uploadFile("doc-19mb.pdf", "application/pdf", "drop", body, when))
		}
		if err != nil {
			t.Fatal(err)
		}
		parts := docupload.SplitBlocks(prepped.Compressed, docupload.DefaultBlockSize)
		if len(parts) < 2 || len(parts[len(parts)-1]) >= docupload.DefaultBlockSize {
			t.Fatalf("blocks = %d, last = %d", len(parts), len(parts[len(parts)-1]))
		}
		started, err := env.client("acct-1").Start(context.Background(), prepped)
		if err != nil {
			t.Fatal(err)
		}
		if len(started.Pending) != len(parts) {
			t.Fatalf("pending = %d, blocks = %d", len(started.Pending), len(parts))
		}
		if err := env.client("acct-1").SendPending(context.Background(), prepped, started); err != nil {
			t.Fatal(err)
		}
		doc, err := env.svc.GetCompleteDoc(context.Background(), "acct-1", prepped.FileID)
		if err != nil {
			t.Fatal(err)
		}
		assertDoc(t, doc, prepped)
		if doc.ContentType != "application/pdf" || doc.Source != "drop" {
			t.Fatalf("meta = %+v", doc)
		}
	})

	t.Run("over20", func(t *testing.T) {
		env := newTestEnv(t, nil)
		_, _, err := env.client("acct-1").Upload(context.Background(), uploadFile("too-big.jpg", "image/jpeg", "gallery", files["too-big.jpg"], when))
		if connectCode(err) != connect.CodeInvalidArgument {
			t.Fatalf("oversize code = %s, err = %v", connectCode(err), err)
		}
		if n := len(env.mrKeys(t)); n != 0 {
			t.Fatalf("redis keys = %d", n)
		}
	})
}

func (e *testEnv) mrKeys(t *testing.T) []string {
	t.Helper()
	keys, err := e.rdb.Keys(context.Background(), "*").Result()
	if err != nil {
		t.Fatal(err)
	}
	return keys
}

func TestClientProtocol(t *testing.T) {
	env := newTestEnv(t, nil)
	client := env.client("acct-1")
	ctx := context.Background()
	when := time.UnixMilli(1_700_000_111_000).UTC()
	body := sizedPDF(docupload.DefaultBlockSize + 100)
	file := uploadFile(`C:\docs\scan.pdf`, "application/pdf", "gallery", body, when)
	prepped, err := docupload.Prepare(file)
	if err != nil {
		t.Fatal(err)
	}
	started, err := client.Start(ctx, prepped)
	if err != nil {
		t.Fatal(err)
	}
	blocks := docupload.SplitBlocks(prepped.Compressed, int(started.BlockSize))
	if len(blocks) != 2 || len(blocks[1]) >= int(started.BlockSize) {
		t.Fatalf("blocks=%d last=%d", len(blocks), len(blocks[len(blocks)-1]))
	}

	t.Run("bad md5", func(t *testing.T) {
		sum := md5Hex(blocks[0])
		bad := "0" + sum[1:]
		if bad == sum {
			bad = "1" + sum[1:]
		}
		got, err := client.PutBlock(ctx, prepped.FileID, 0, blocks[0], bad)
		if err != nil || got.Success || !containsIndex(got.Pending, 0) {
			t.Fatalf("bad md5 result=%+v err=%v", got, err)
		}
	})

	t.Run("length", func(t *testing.T) {
		short := blocks[0][:len(blocks[0])-1]
		got, err := client.PutBlock(ctx, prepped.FileID, 0, short, "")
		if err != nil || got.Success || !containsIndex(got.Pending, 0) {
			t.Fatalf("short result=%+v err=%v", got, err)
		}
		long := append(append([]byte{}, blocks[0]...), 0x7f)
		got, err = client.PutBlock(ctx, prepped.FileID, 0, long, "")
		if err != nil || got.Success || !containsIndex(got.Pending, 0) {
			t.Fatalf("long result=%+v err=%v", got, err)
		}
	})

	t.Run("index and unknown", func(t *testing.T) {
		got, err := client.PutBlock(ctx, prepped.FileID, 9, blocks[0], "")
		if err != nil || got.Success {
			t.Fatalf("index result=%+v err=%v", got, err)
		}
		_, err = client.PutBlock(ctx, strings.Repeat("ab", 16), 0, blocks[0], "")
		if connectCode(err) != connect.CodeNotFound {
			t.Fatalf("unknown code=%s err=%v", connectCode(err), err)
		}
	})

	t.Run("duplicate and conflict", func(t *testing.T) {
		first, err := client.PutBlock(ctx, prepped.FileID, 0, blocks[0], "")
		if err != nil || !first.Success || containsIndex(first.Pending, 0) {
			t.Fatalf("first=%+v err=%v", first, err)
		}
		again, err := client.PutBlock(ctx, prepped.FileID, 0, blocks[0], "")
		if err != nil || !again.Success {
			t.Fatalf("duplicate=%+v err=%v", again, err)
		}
		other := bytes.Clone(blocks[0])
		other[0] ^= 0xff
		_, err = client.PutBlock(ctx, prepped.FileID, 0, other, md5Hex(other))
		if connectCode(err) != connect.CodeAlreadyExists {
			t.Fatalf("conflict code=%s err=%v", connectCode(err), err)
		}
	})

	t.Run("out of order", func(t *testing.T) {
		got, err := client.PutBlock(ctx, prepped.FileID, 1, blocks[1], "")
		if err != nil || !got.Success || len(got.Pending) != 0 {
			t.Fatalf("tail=%+v err=%v", got, err)
		}
		doc, err := env.svc.GetCompleteDoc(ctx, "acct-1", prepped.FileID)
		if err != nil {
			t.Fatal(err)
		}
		assertDoc(t, doc, prepped)
		if doc.FileName != "scan.pdf" || doc.Source != "gallery" {
			t.Fatalf("cleaned meta = %+v", doc)
		}
	})
}

func TestClientResumeResetCancelIsolationExpiry(t *testing.T) {
	ctx := context.Background()
	when := time.UnixMilli(1_711_000_000_000).UTC()
	body := sizedPDF(docupload.DefaultBlockSize + 80)

	t.Run("resume", func(t *testing.T) {
		env := newTestEnv(t, nil)
		client := env.client("acct-1")
		prepped, err := docupload.Prepare(uploadFile("resume.pdf", "application/pdf", "drop", body, when))
		if err != nil {
			t.Fatal(err)
		}
		started, err := client.Start(ctx, prepped)
		if err != nil {
			t.Fatal(err)
		}
		blocks := docupload.SplitBlocks(prepped.Compressed, int(started.BlockSize))
		if _, err := client.PutBlock(ctx, prepped.FileID, 0, blocks[0], ""); err != nil {
			t.Fatal(err)
		}
		_, err = env.svc.GetCompleteDoc(ctx, "acct-1", prepped.FileID)
		if appCode(err) != docupload.ErrIncomplete.Code {
			t.Fatalf("incomplete code=%s err=%v", appCode(err), err)
		}
		again, err := client.Start(ctx, prepped)
		if err != nil {
			t.Fatal(err)
		}
		if containsIndex(again.Pending, 0) || !containsIndex(again.Pending, 1) {
			t.Fatalf("resume pending=%v", again.Pending)
		}
		if err := client.SendPending(ctx, prepped, again); err != nil {
			t.Fatal(err)
		}
		doc, err := env.svc.GetCompleteDoc(ctx, "acct-1", prepped.FileID)
		if err != nil {
			t.Fatal(err)
		}
		assertDoc(t, doc, prepped)
	})

	t.Run("fingerprint reset", func(t *testing.T) {
		env := newTestEnv(t, nil)
		client := env.client("acct-1")
		prepped, err := docupload.Prepare(uploadFile("reset.pdf", "application/pdf", "drop", body, when))
		if err != nil {
			t.Fatal(err)
		}
		started, err := client.Start(ctx, prepped)
		if err != nil {
			t.Fatal(err)
		}
		blocks := docupload.SplitBlocks(prepped.Compressed, int(started.BlockSize))
		if _, err := client.PutBlock(ctx, prepped.FileID, 0, blocks[0], ""); err != nil {
			t.Fatal(err)
		}
		alt, err := gzipNoCompression(body)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Equal(alt, prepped.Compressed) {
			t.Fatal("alternate gzip matched the first body")
		}
		replaced := prepped
		replaced.Compressed = alt
		replaced.CompressedMD5 = md5Hex(alt)
		next, err := client.Start(ctx, replaced)
		if err != nil {
			t.Fatal(err)
		}
		altBlocks := docupload.SplitBlocks(alt, int(next.BlockSize))
		if len(next.Pending) != len(altBlocks) {
			t.Fatalf("reset pending=%d blocks=%d", len(next.Pending), len(altBlocks))
		}
		if err := client.SendPending(ctx, replaced, next); err != nil {
			t.Fatal(err)
		}
		doc, err := env.svc.GetCompleteDoc(ctx, "acct-1", replaced.FileID)
		if err != nil {
			t.Fatal(err)
		}
		assertDoc(t, doc, replaced)
	})

	t.Run("cancel", func(t *testing.T) {
		env := newTestEnv(t, nil)
		client := env.client("acct-1")
		prepped, started, err := client.Upload(ctx, uploadFile("cancel.pdf", "application/pdf", "drop", smallJPEG(), when))
		if err != nil {
			t.Fatal(err)
		}
		_ = started
		if err := client.Cancel(ctx, prepped.FileID); err != nil {
			t.Fatal(err)
		}
		_, err = env.svc.GetCompleteDoc(ctx, "acct-1", prepped.FileID)
		if appCode(err) != docupload.ErrNotFound.Code {
			t.Fatalf("after cancel code=%s err=%v", appCode(err), err)
		}
	})

	t.Run("other account", func(t *testing.T) {
		env := newTestEnv(t, nil)
		owner := env.client("acct-1")
		prepped, err := docupload.Prepare(uploadFile("private.pdf", "application/pdf", "drop", body, when))
		if err != nil {
			t.Fatal(err)
		}
		started, err := owner.Start(ctx, prepped)
		if err != nil {
			t.Fatal(err)
		}
		blocks := docupload.SplitBlocks(prepped.Compressed, int(started.BlockSize))
		_, err = env.client("acct-2").PutBlock(ctx, prepped.FileID, 0, blocks[0], "")
		if connectCode(err) != connect.CodeNotFound {
			t.Fatalf("other account code=%s err=%v", connectCode(err), err)
		}
		_, err = env.svc.GetCompleteDoc(ctx, "acct-2", prepped.FileID)
		if appCode(err) != docupload.ErrNotFound.Code {
			t.Fatalf("other get code=%s err=%v", appCode(err), err)
		}
		if err := owner.SendPending(ctx, prepped, started); err != nil {
			t.Fatal(err)
		}
		doc, err := env.svc.GetCompleteDoc(ctx, "acct-1", prepped.FileID)
		if err != nil {
			t.Fatal(err)
		}
		assertDoc(t, doc, prepped)
	})

	t.Run("expiry and sweep", func(t *testing.T) {
		env := newTestEnv(t, nil)
		client := env.client("acct-1")
		prepped, err := docupload.Prepare(uploadFile("expire.pdf", "application/pdf", "drop", body, when))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.Start(ctx, prepped); err != nil {
			t.Fatal(err)
		}
		past := time.Now().Add(-time.Second).UnixMilli()
		if err := env.rdb.HSet(ctx, docupload.MetaKey("acct-1", prepped.FileID), "expires_unix_ms", strconv.FormatInt(past, 10)).Err(); err != nil {
			t.Fatal(err)
		}
		blocks := docupload.SplitBlocks(prepped.Compressed, docupload.DefaultBlockSize)
		_, err = client.PutBlock(ctx, prepped.FileID, 0, blocks[0], "")
		if connectCode(err) != connect.CodeFailedPrecondition {
			t.Fatalf("expired code=%s err=%v", connectCode(err), err)
		}
		removed, err := env.svc.Sweep(ctx, time.Now())
		if err != nil || removed < 1 {
			t.Fatalf("sweep removed=%d err=%v", removed, err)
		}
		for _, key := range env.mrKeys(t) {
			if strings.Contains(key, prepped.FileID) {
				t.Fatalf("session key remains: %s", key)
			}
		}
		orphan := "docupload:acct-1:" + strings.Repeat("cd", 16) + ":blk:0"
		if err := env.rdb.Set(ctx, orphan, "x", 0).Err(); err != nil {
			t.Fatal(err)
		}
		if _, err := env.svc.Sweep(ctx, time.Now()); err != nil {
			t.Fatal(err)
		}
		if n, err := env.rdb.Exists(ctx, orphan).Result(); err != nil || n != 0 {
			t.Fatalf("orphan exists=%d err=%v", n, err)
		}
		if _, err := client.Start(ctx, prepped); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("expiry clamp", func(t *testing.T) {
		env := newTestEnv(t, nil)
		file := uploadFile("clamp.pdf", "application/pdf", "drop", smallJPEG(), when)
		file.ExpiresAt = time.Now().Add(365 * 24 * time.Hour)
		prepped, err := docupload.Prepare(file)
		if err != nil {
			t.Fatal(err)
		}
		started, err := env.client("acct-1").Start(ctx, prepped)
		if err != nil {
			t.Fatal(err)
		}
		max := time.Now().Add(docupload.MaxTTL)
		got := time.UnixMilli(started.ExpiresUnixMs)
		if got.Before(max.Add(-time.Minute)) || got.After(max.Add(time.Minute)) {
			t.Fatalf("clamped expiry %s, max %s", got, max)
		}
		file.ExpiresAt = time.Now().Add(-time.Hour)
		prepped, err = docupload.Prepare(file)
		if err != nil {
			t.Fatal(err)
		}
		_, err = env.client("acct-1").Start(ctx, prepped)
		if connectCode(err) != connect.CodeFailedPrecondition {
			t.Fatalf("past expiry code=%s err=%v", connectCode(err), err)
		}
	})
}

func TestAssembleTamperAndBomb(t *testing.T) {
	ctx := context.Background()
	env := newTestEnv(t, nil)
	when := time.UnixMilli(1_722_000_000_000).UTC()
	client := env.client("acct-1")
	prepped, _, err := client.Upload(ctx, uploadFile("photo.jpg", "image/jpeg", "camera", smallJPEG(), when))
	if err != nil {
		t.Fatal(err)
	}
	key := docupload.BlockObjectKey("acct-1", prepped.FileID, 0)
	raw, err := env.rdb.Get(ctx, key).Bytes()
	if err != nil {
		t.Fatal(err)
	}
	raw[0] ^= 0xff
	if err := env.rdb.Set(ctx, key, raw, 0).Err(); err != nil {
		t.Fatal(err)
	}
	_, err = env.svc.GetCompleteDoc(ctx, "acct-1", prepped.FileID)
	if appCode(err) != docupload.ErrChecksum.Code {
		t.Fatalf("tamper code=%s err=%v", appCode(err), err)
	}

	plain := bytes.Repeat([]byte{0}, int(docupload.DefaultMaxOriginalBytes)+1)
	compressed, err := docupload.Gzip(plain)
	if err != nil {
		t.Fatal(err)
	}
	bomb := docupload.Prepared{
		File: docupload.UploadFile{
			Name:        "bomb.bin",
			Bytes:       make([]byte, 32),
			ContentType: "application/octet-stream",
			Source:      "drop",
			FileTime:    when,
		},
		FileID:        md5Hex(plain),
		Compressed:    compressed,
		CompressedMD5: md5Hex(compressed),
		FileUnixMs:    when.UnixMilli(),
	}
	started, err := client.Start(ctx, bomb)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.SendPending(ctx, bomb, started); err != nil {
		t.Fatal(err)
	}
	_, err = env.svc.GetCompleteDoc(ctx, "acct-1", bomb.FileID)
	if appCode(err) != docupload.ErrTooLarge.Code {
		t.Fatalf("bomb code=%s err=%v", appCode(err), err)
	}
}

func TestClientRetriesTransientFailures(t *testing.T) {
	var left atomic.Int32
	left.Store(2)
	env := newTestEnv(t, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "UploadBlock") && left.Add(-1) >= 0 {
				http.Error(w, "temporary", http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	when := time.UnixMilli(1_733_000_000_000).UTC()
	prepped, _, err := env.client("acct-1").Upload(context.Background(), uploadFile("photo.jpg", "image/jpeg", "camera", smallJPEG(), when))
	if err != nil {
		t.Fatal(err)
	}
	doc, err := env.svc.GetCompleteDoc(context.Background(), "acct-1", prepped.FileID)
	if err != nil {
		t.Fatal(err)
	}
	assertDoc(t, doc, prepped)
}
