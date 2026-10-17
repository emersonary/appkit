package docupload

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const putBlockLua = `
local exists = redis.call('EXISTS', KEYS[1])
if exists == 0 then
  return {'missing', '0', '0'}
end

local expires = tonumber(redis.call('HGET', KEYS[1], 'expires_unix_ms'))
local now = tonumber(ARGV[3])
if expires == nil or now == nil or expires < now then
  return {'expired', '0', '0'}
end

local count = tonumber(redis.call('HGET', KEYS[1], 'block_count'))
local blockSize = tonumber(redis.call('HGET', KEYS[1], 'block_size'))
local compressed = tonumber(redis.call('HGET', KEYS[1], 'compressed_size'))
local index = tonumber(ARGV[1])
if count == nil or blockSize == nil or compressed == nil or index == nil or index < 0 or index >= count then
  return {'index', '0', '0'}
end

local expected = blockSize
if index == count - 1 then
  expected = compressed - (count - 1) * blockSize
end
if string.len(ARGV[2]) ~= expected then
  return {'length', '0', '0'}
end

local prev = redis.call('GET', KEYS[3])
local added = 0
if prev then
  if prev ~= ARGV[2] then
    return {'conflict', '0', '0'}
  end
else
  redis.call('SET', KEYS[3], ARGV[2])
  added = redis.call('SADD', KEYS[2], ARGV[1])
end

local ttl = tonumber(ARGV[4])
if ttl == nil or ttl < 1 then
  ttl = 1
end
redis.call('EXPIRE', KEYS[1], ttl)
redis.call('EXPIRE', KEYS[2], ttl)
local prefix = ARGV[5]
for i = 0, count - 1 do
  redis.call('EXPIRE', prefix .. i, ttl)
end

local n = redis.call('SCARD', KEYS[2])
return {'ok', tostring(added), tostring(n)}
`

// Store keeps upload sessions in Redis.
type Store struct {
	rdb    *redis.Client
	script *redis.Script
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{rdb: rdb, script: redis.NewScript(putBlockLua)}
}

func metaKey(account, fileID string) string {
	return fmt.Sprintf("docupload:%s:%s:meta", account, fileID)
}

func recvKey(account, fileID string) string {
	return fmt.Sprintf("docupload:%s:%s:recv", account, fileID)
}

func blockKey(account, fileID string, index int32) string {
	return fmt.Sprintf("docupload:%s:%s:blk:%d", account, fileID, index)
}

func blockPrefix(account, fileID string) string {
	return fmt.Sprintf("docupload:%s:%s:blk:", account, fileID)
}

// MetaKey is the Redis hash that holds one upload session.
func MetaKey(account, fileID string) string { return metaKey(account, fileID) }

// BlockObjectKey is the Redis key for one compressed slice.
func BlockObjectKey(account, fileID string, index int32) string {
	return blockKey(account, fileID, index)
}

func sessionPattern(account, fileID string) string {
	return fmt.Sprintf("docupload:%s:%s:*", account, fileID)
}

func (s *Store) SaveNew(ctx context.Context, session Session, ttl time.Duration) error {
	if err := s.deletePattern(ctx, sessionPattern(session.AccountID, session.FileID)); err != nil {
		return err
	}
	meta := metaKey(session.AccountID, session.FileID)
	fields := map[string]any{
		"account_id":      session.AccountID,
		"file_id":         session.FileID,
		"file_name":       session.FileName,
		"file_unix_ms":    strconv.FormatInt(session.FileUnixMs, 10),
		"original_size":   strconv.FormatInt(session.OriginalSize, 10),
		"compressed_md5":  session.CompressedMD5,
		"compressed_size": strconv.FormatInt(session.CompressedSize, 10),
		"block_size":      strconv.Itoa(session.BlockSize),
		"block_count":     strconv.Itoa(session.BlockCount),
		"max_threads":     strconv.Itoa(session.MaxThreads),
		"expires_unix_ms": strconv.FormatInt(session.ExpiresUnixMs, 10),
		"content_type":    session.ContentType,
		"source":          session.Source,
	}
	pipe := s.rdb.Pipeline()
	pipe.HSet(ctx, meta, fields)
	sec := ttlDuration(ttl)
	pipe.Expire(ctx, meta, sec)
	pipe.Expire(ctx, recvKey(session.AccountID, session.FileID), sec)
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Store) Get(ctx context.Context, account, fileID string) (Session, error) {
	vals, err := s.rdb.HGetAll(ctx, metaKey(account, fileID)).Result()
	if err != nil {
		return Session{}, err
	}
	if len(vals) == 0 {
		return Session{}, ErrNotFound
	}
	session, err := sessionFromHash(vals)
	if err != nil {
		return Session{}, err
	}
	if session.ExpiresUnixMs < time.Now().UnixMilli() {
		return session, ErrExpired
	}
	return session, nil
}

func (s *Store) Touch(ctx context.Context, session Session, expiresUnixMs int64, ttl time.Duration) error {
	sec := ttlDuration(ttl)
	pipe := s.rdb.Pipeline()
	meta := metaKey(session.AccountID, session.FileID)
	pipe.HSet(ctx, meta, "expires_unix_ms", strconv.FormatInt(expiresUnixMs, 10))
	pipe.Expire(ctx, meta, sec)
	pipe.Expire(ctx, recvKey(session.AccountID, session.FileID), sec)
	for i := 0; i < session.BlockCount; i++ {
		pipe.Expire(ctx, blockKey(session.AccountID, session.FileID, int32(i)), sec)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) Delete(ctx context.Context, account, fileID string) error {
	return s.deletePattern(ctx, sessionPattern(account, fileID))
}

// PutBlock stores one slice atomically. status is ok, missing, expired, index, length, or conflict.
// added is 1 when this call inserted a new index. received is the set size after an ok write.
func (s *Store) PutBlock(ctx context.Context, account, fileID string, index int32, payload []byte, now time.Time, ttl time.Duration) (status string, added int, received int, err error) {
	res, err := s.script.Run(ctx, s.rdb, []string{
		metaKey(account, fileID),
		recvKey(account, fileID),
		blockKey(account, fileID, index),
	}, strconv.FormatInt(int64(index), 10), string(payload), strconv.FormatInt(now.UnixMilli(), 10), strconv.Itoa(ttlSecs(ttl)), blockPrefix(account, fileID)).Result()
	if err != nil {
		return "", 0, 0, err
	}
	return parseScript(res)
}

func (s *Store) Received(ctx context.Context, account, fileID string) (map[int32]struct{}, error) {
	members, err := s.rdb.SMembers(ctx, recvKey(account, fileID)).Result()
	if err != nil {
		return nil, err
	}
	got := make(map[int32]struct{}, len(members))
	for _, member := range members {
		n, err := strconv.ParseInt(member, 10, 32)
		if err != nil {
			return nil, err
		}
		got[int32(n)] = struct{}{}
	}
	return got, nil
}

func (s *Store) Block(ctx context.Context, account, fileID string, index int32) ([]byte, error) {
	b, err := s.rdb.Get(ctx, blockKey(account, fileID, index)).Bytes()
	if err == redis.Nil {
		return nil, ErrIncomplete
	}
	return b, err
}

func (s *Store) SetExpires(ctx context.Context, account, fileID string, expiresUnixMs int64) error {
	return s.rdb.HSet(ctx, metaKey(account, fileID), "expires_unix_ms", strconv.FormatInt(expiresUnixMs, 10)).Err()
}

// Sweep deletes sessions whose expiry has passed and block keys whose meta is gone.
func (s *Store) Sweep(ctx context.Context, now time.Time) (int, error) {
	removed := 0
	var cursor uint64
	for {
		keys, next, err := s.rdb.Scan(ctx, cursor, "docupload:*:meta", 128).Result()
		if err != nil {
			return removed, err
		}
		for _, key := range keys {
			raw, err := s.rdb.HGet(ctx, key, "expires_unix_ms").Result()
			if err == redis.Nil {
				continue
			}
			if err != nil {
				return removed, err
			}
			expires, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return removed, err
			}
			if expires >= now.UnixMilli() {
				continue
			}
			account, fileID, ok := splitMetaKey(key)
			if !ok {
				continue
			}
			if err := s.Delete(ctx, account, fileID); err != nil {
				return removed, err
			}
			removed++
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}

	cursor = 0
	for {
		keys, next, err := s.rdb.Scan(ctx, cursor, "docupload:*:blk:*", 128).Result()
		if err != nil {
			return removed, err
		}
		for _, key := range keys {
			account, fileID, ok := splitBlockKey(key)
			if !ok {
				continue
			}
			n, err := s.rdb.Exists(ctx, metaKey(account, fileID)).Result()
			if err != nil {
				return removed, err
			}
			if n == 0 {
				if err := s.rdb.Del(ctx, key).Err(); err != nil {
					return removed, err
				}
				removed++
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return removed, nil
}

func (s *Store) deletePattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, next, err := s.rdb.Scan(ctx, cursor, pattern, 128).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := s.rdb.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			return nil
		}
	}
}

func sessionFromHash(vals map[string]string) (Session, error) {
	fileUnix, err := strconv.ParseInt(vals["file_unix_ms"], 10, 64)
	if err != nil {
		return Session{}, err
	}
	original, err := strconv.ParseInt(vals["original_size"], 10, 64)
	if err != nil {
		return Session{}, err
	}
	compressed, err := strconv.ParseInt(vals["compressed_size"], 10, 64)
	if err != nil {
		return Session{}, err
	}
	blockSize, err := strconv.Atoi(vals["block_size"])
	if err != nil {
		return Session{}, err
	}
	blockCount, err := strconv.Atoi(vals["block_count"])
	if err != nil {
		return Session{}, err
	}
	maxThreads, err := strconv.Atoi(vals["max_threads"])
	if err != nil {
		return Session{}, err
	}
	expires, err := strconv.ParseInt(vals["expires_unix_ms"], 10, 64)
	if err != nil {
		return Session{}, err
	}
	return Session{
		AccountID:      vals["account_id"],
		FileID:         vals["file_id"],
		FileName:       vals["file_name"],
		FileUnixMs:     fileUnix,
		OriginalSize:   original,
		CompressedMD5:  vals["compressed_md5"],
		CompressedSize: compressed,
		BlockSize:      blockSize,
		BlockCount:     blockCount,
		MaxThreads:     maxThreads,
		ExpiresUnixMs:  expires,
		ContentType:    vals["content_type"],
		Source:         vals["source"],
	}, nil
}

func splitMetaKey(key string) (account, fileID string, ok bool) {
	const prefix = "docupload:"
	const suffix = ":meta"
	if len(key) <= len(prefix)+len(suffix) || key[:len(prefix)] != prefix || key[len(key)-len(suffix):] != suffix {
		return "", "", false
	}
	rest := key[len(prefix) : len(key)-len(suffix)]
	for i := 0; i < len(rest); i++ {
		if rest[i] == ':' {
			return rest[:i], rest[i+1:], rest[:i] != "" && rest[i+1:] != ""
		}
	}
	return "", "", false
}

func splitBlockKey(key string) (account, fileID string, ok bool) {
	const prefix = "docupload:"
	const marker = ":blk:"
	if len(key) <= len(prefix) || key[:len(prefix)] != prefix {
		return "", "", false
	}
	rest := key[len(prefix):]
	idx := -1
	for i := 0; i+len(marker) <= len(rest); i++ {
		if rest[i:i+len(marker)] == marker {
			idx = i
			break
		}
	}
	if idx <= 0 {
		return "", "", false
	}
	head := rest[:idx]
	for i := 0; i < len(head); i++ {
		if head[i] == ':' {
			return head[:i], head[i+1:], head[:i] != "" && head[i+1:] != ""
		}
	}
	return "", "", false
}

func parseScript(res any) (string, int, int, error) {
	var parts []string
	switch vals := res.(type) {
	case []any:
		for _, v := range vals {
			parts = append(parts, fmt.Sprint(v))
		}
	case []string:
		parts = vals
	default:
		return "", 0, 0, fmt.Errorf("unexpected script result %T", res)
	}
	if len(parts) < 3 {
		return "", 0, 0, fmt.Errorf("short script result")
	}
	added, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, 0, err
	}
	received, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", 0, 0, err
	}
	return parts[0], added, received, nil
}

func ttlSecs(ttl time.Duration) int {
	sec := int(ttl.Seconds())
	if sec < 1 {
		return 1
	}
	return sec
}

func ttlDuration(ttl time.Duration) time.Duration {
	return time.Duration(ttlSecs(ttl)) * time.Second
}
