package docupload

import "time"

const (
	DefaultBlockSize          = 256 * 1024
	DefaultMaxThreads         = 4
	DefaultTTL                = 24 * time.Hour
	MaxTTL                    = 72 * time.Hour
	DefaultMaxOriginalBytes   = 20 << 20
	DefaultMaxCompressedBytes = 20 << 20
)

// Config controls block sizing, concurrency, and session lifetime.
type Config struct {
	BlockSize          int
	MaxThreads         int
	DefaultTTL         time.Duration
	MaxTTL             time.Duration
	MaxOriginalBytes   int64
	MaxCompressedBytes int64
}

// DefaultConfig is the production upload policy.
func DefaultConfig() Config {
	return Config{
		BlockSize:          DefaultBlockSize,
		MaxThreads:         DefaultMaxThreads,
		DefaultTTL:         DefaultTTL,
		MaxTTL:             MaxTTL,
		MaxOriginalBytes:   DefaultMaxOriginalBytes,
		MaxCompressedBytes: DefaultMaxCompressedBytes,
	}
}

func (c Config) normalized() Config {
	out := c
	def := DefaultConfig()
	if out.BlockSize <= 0 {
		out.BlockSize = def.BlockSize
	}
	if out.MaxThreads <= 0 {
		out.MaxThreads = def.MaxThreads
	}
	if out.DefaultTTL <= 0 {
		out.DefaultTTL = def.DefaultTTL
	}
	if out.MaxTTL <= 0 {
		out.MaxTTL = def.MaxTTL
	}
	if out.MaxTTL < out.DefaultTTL {
		out.MaxTTL = out.DefaultTTL
	}
	if out.MaxOriginalBytes <= 0 {
		out.MaxOriginalBytes = def.MaxOriginalBytes
	}
	if out.MaxCompressedBytes <= 0 {
		out.MaxCompressedBytes = def.MaxCompressedBytes
	}
	return out
}
