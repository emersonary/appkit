// Package uuidv7 implements UUID version 7 generation and parsing per
// github.com/emersonary/uuidv7 (millisecond timestamp + sub-ms entropy in rand bits).
package uuidv7

import (
	"crypto/rand"
	"time"

	"github.com/google/uuid"
)

func putUint48(b []byte, millis, nanos uint64) {
	_ = b[15] // bounds check
	b[0] = byte(millis >> 40)
	b[1] = byte(millis >> 32)
	b[2] = byte(millis >> 24)
	b[3] = byte(millis >> 16)
	b[4] = byte(millis >> 8)
	b[5] = byte(millis)

	b[6] = byte(nanos >> 16)
	b[7] = byte(nanos >> 8)
	byt := byte(nanos)

	var rnd [2]byte
	_, _ = rand.Read(rnd[:])

	b[8] = (b[8] & 0x0F) | (byt >> 4)
	b[9] = (rnd[0] & 0x0F) | (byt << 4)

	var tail [6]byte
	_, _ = rand.Read(tail[:])
	copy(b[10:16], tail[:])
}

func uint48(b []byte) uint64 {
	return uint64(b[5]) |
		uint64(b[4])<<8 |
		uint64(b[3])<<16 |
		uint64(b[2])<<24 |
		uint64(b[1])<<32 |
		uint64(b[0])<<40
}

func uint24(b []byte) uint64 {
	return uint64(b[2]) |
		uint64(b[1])<<8 |
		uint64(b[0])<<16
}

// NewUUIDv7 returns a new UUID v7 based on the current time.
func NewUUIDv7() uuid.UUID {
	return UUIDv7FromTimeStamp(time.Now())
}

// UUIDv7FromTimeStamp builds a UUID v7 from timeStamp.
func UUIDv7FromTimeStamp(timeStamp time.Time) uuid.UUID {
	unixTimeMilli := timeStamp.UnixMilli()
	unixTimeNano := timeStamp.UnixNano()
	nanosAfterMilli := unixTimeNano - unixTimeMilli*1_000_000

	var id uuid.UUID
	putUint48(id[:], uint64(unixTimeMilli), uint64(nanosAfterMilli))
	id[6] = (id[6] & 0x0F) | 0x70
	id[8] = (id[8] & 0x3F) | 0x80
	return id
}

// TimeStampFromUUIDv7 extracts the embedded timestamp from a UUID v7.
func TimeStampFromUUIDv7(id uuid.UUID) time.Time {
	timestampMilli := uint48(id[0:6])
	byt := make([]byte, 3)
	byt[0] = id[6] & 0x0F
	byt[1] = id[7]
	byt[2] = (id[8]&0x0F)<<4 + id[9]>>4
	timeStampNano := uint24(byt)
	return time.UnixMilli(int64(timestampMilli)).Add(time.Nanosecond * time.Duration(timeStampNano))
}

// MustParse parses s as a UUID (any version).
func MustParse(s string) uuid.UUID {
	return uuid.MustParse(s)
}

// Version returns the UUID version nibble.
func Version(id uuid.UUID) int {
	return int(id[6] >> 4)
}

// MillisFromUUIDv7 returns the 48-bit unix millis component.
func MillisFromUUIDv7(id uuid.UUID) int64 {
	return int64(uint48(id[0:6]))
}
