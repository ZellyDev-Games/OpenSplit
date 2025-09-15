package session

import (
	"time"

	"github.com/google/uuid"
)

type RunPayload struct {
	ID               uuid.UUID      `json:"id"`
	SplitFileVersion int            `json:"splitfile_version"`
	TotalTime        time.Duration  `json:"total_time"`
	Completed        bool           `json:"completed"`
	SplitPayloads    []SplitPayload `json:"split_payloads"`
}

// Run maintains the history attempts in a SplitFile.
//
// This is useful to calculate things like gold splits, sum of best segments, etc...
type Run struct {
	id               uuid.UUID
	splitFileVersion int
	totalTime        time.Duration
	completed        bool
	splitPayloads    []SplitPayload
}

// GetPayload returns a snapshot of a Run
//
// GetPayload, modify the payload, then send it to NewRunFromPayload and persist the result of that to make changes.
func (r *Run) GetPayload() RunPayload {
	return RunPayload{
		ID:               r.id,
		SplitFileVersion: r.splitFileVersion,
		TotalTime:        r.totalTime,
		Completed:        r.completed,
		SplitPayloads:    r.splitPayloads,
	}
}

// NewRunFromPayload creates a new Run from a RunPayload.
//
// Useful for making stateful updates to a Run without exposing internal data structure or presentation.
func NewRunFromPayload(payload RunPayload) Run {
	return Run{
		id:               payload.ID,
		splitFileVersion: payload.SplitFileVersion,
		totalTime:        payload.TotalTime,
		completed:        payload.Completed,
		splitPayloads:    payload.SplitPayloads,
	}
}
