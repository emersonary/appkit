package docupload

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"io"
	"time"
)

func md5Hex(b []byte) string {
	sum := md5.Sum(b)
	return hex.EncodeToString(sum[:])
}

// Gzip compresses b with a zero mtime so the same input yields the same bytes.
func Gzip(b []byte) ([]byte, error) {
	return gzipLevel(b, gzip.DefaultCompression)
}

func gzipLevel(b []byte, level int) ([]byte, error) {
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, level)
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

func gunzipLimited(src []byte, limit int64) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	out, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(out)) > limit {
		return nil, ErrTooLarge
	}
	return out, nil
}

// SplitBlocks slices data into pieces of size, with a shorter tail.
func SplitBlocks(data []byte, size int) [][]byte {
	if size <= 0 || len(data) == 0 {
		return nil
	}
	out := make([][]byte, 0, (len(data)+size-1)/size)
	for i := 0; i < len(data); i += size {
		j := i + size
		if j > len(data) {
			j = len(data)
		}
		out = append(out, data[i:j])
	}
	return out
}

func blockCountOf(size int64, blockSize int) int {
	bs := int64(blockSize)
	if size <= 0 || bs <= 0 {
		return 0
	}
	return int((size + bs - 1) / bs)
}

func pendingIndexes(count int, got map[int32]struct{}) []int32 {
	pending := make([]int32, 0, count)
	for i := 0; i < count; i++ {
		idx := int32(i)
		if _, ok := got[idx]; !ok {
			pending = append(pending, idx)
		}
	}
	return pending
}

func allIndexes(count int) []int32 {
	out := make([]int32, count)
	for i := 0; i < count; i++ {
		out[i] = int32(i)
	}
	return out
}
