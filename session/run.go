package session

import (
	"time"

	"github.com/google/uuid"
)

type RunPayload struct {
	ID               uuid.UUID      `json:"id"`
	SplitFileVersion int            `json:"splitFileVersion"`
	TotalTime        time.Duration  `json:"totalTime"`
	Completed        bool           `json:"completed"`
	SplitPayloads    []SplitPayload `json:"splitPayloads"`
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

func (r *Run) GetPayload() RunPayload {
	return RunPayload{
		ID:               r.id,
		SplitFileVersion: r.splitFileVersion,
		TotalTime:        r.totalTime,
		Completed:        r.completed,
		SplitPayloads:    r.splitPayloads,
	}
}

func NewRunFromPayload(payload RunPayload) Run {
	return Run{
		id:               payload.ID,
		splitFileVersion: payload.SplitFileVersion,
		totalTime:        payload.TotalTime,
		completed:        payload.Completed,
		splitPayloads:    payload.SplitPayloads,
	}
}
