package session

import (
	"time"

	"github.com/google/uuid"
)

// SplitPayload is a snapshot of split data to communicate information about a split to the frontend, and also the
// Run history in SplitFile runs
type SplitPayload struct {
	SplitIndex     int      `json:"split_index"`
	SplitSegmentID string   `json:"split_segment_id"`
	CurrentTime    StatTime `json:"current_time"`
}

// Split represents an advancement of a run through the segments.
//
// Split identifies a completed segment, and how long that segment took
type Split struct {
	splitIndex      int
	splitSegmentID  uuid.UUID
	currentDuration time.Duration
}

func (s *Split) getPayload() SplitPayload {
	return SplitPayload{
		SplitIndex:     s.splitIndex,
		SplitSegmentID: s.splitSegmentID.String(),
		CurrentTime: StatTime{
			Raw:       s.currentDuration.Milliseconds(),
			Formatted: FormatTimeToString(s.currentDuration),
		},
	}
}
