// Package uuid is the platform entry point for generating primary-key UUIDs.
// All application and repo code should use this package instead of github.com/google/uuid
// or Postgres gen_random_uuid().
package uuid

import (
	"time"

	"github.com/emersonary/appkit/uuidv7"
	"github.com/google/uuid"
)

// New returns a time-ordered UUID v7 for the current timestamp.
func New() uuid.UUID {
	return uuidv7.NewUUIDv7()
}

// NewString returns New().String().
func NewString() string {
	return uuidv7.NewUUIDv7().String()
}

// FromTime returns a UUID v7 embedding timeStamp.
func FromTime(timeStamp time.Time) uuid.UUID {
	return uuidv7.UUIDv7FromTimeStamp(timeStamp)
}

// TimeFrom extracts the timestamp from a UUID v7.
func TimeFrom(id uuid.UUID) time.Time {
	return uuidv7.TimeStampFromUUIDv7(id)
}
