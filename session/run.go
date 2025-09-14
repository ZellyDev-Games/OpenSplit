package session

import (
	"time"

	"github.com/google/uuid"
)

// Run maintains the history attempts in a SplitFile.
//
// This is useful to calculate things like gold splits, sum of best segments, etc...
type Run struct {
	id               uuid.UUID
	splitFileID      uuid.UUID
	splitFileVersion int
	duration         time.Duration
	completed        bool
	splits           []SplitPayload
}
