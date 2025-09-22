package session

import (
	"time"

	"github.com/google/uuid"
)

// SegmentPayload is a snapshot of a given Segment useful for communicating state while protecting internal data.
type SegmentPayload struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Gold    StatTime `json:"gold"`
	Average StatTime `json:"average"`
	PB      StatTime `json:"pb"`
}

// Segment represents a portion of a game that you want to time (e.g. "Level 1")
type Segment struct {
	id      uuid.UUID
	name    string
	gold    time.Duration
	average time.Duration
	pb      time.Duration
}

// NewFromPayload creates a new Segment from the given SegmentPayload.
//
// This is a pattern used often in OpenSplit where you fetch a payload, modify it, then pass it into a modification or
// creation func to persist changes internally.
func NewFromPayload(payload SegmentPayload) Segment {
	if payload.ID == "" {
		payload.ID = uuid.New().String()
	}

	return Segment{
		id:      uuid.MustParse(payload.ID),
		name:    payload.Name,
		average: time.Duration(payload.Average.Raw) * time.Millisecond,
		pb:      time.Duration(payload.PB.Raw) * time.Millisecond,
	}
}

// GetPayload retrieves a SegmentPayload representing the state of the Segment
func (s *Segment) GetPayload() SegmentPayload {
	return SegmentPayload{
		ID:   s.id.String(),
		Name: s.name,
		Average: StatTime{
			s.average.Milliseconds(),
			FormatTimeToString(s.average),
		},
		PB: StatTime{
			s.pb.Milliseconds(),
			FormatTimeToString(s.pb),
		},
		Gold: StatTime{
			s.gold.Milliseconds(),
			FormatTimeToString(s.gold),
		},
	}
}
