package session

import (
	"time"

	"github.com/google/uuid"
)

type RunPayload struct {
	ID               uuid.UUID `json:"id"`
	SplitFileVersion int       `json:"splitFileVersion"`
}

// Run maintains the history attempts in a SplitFile.
//
// This is useful to calculate things like gold splits, sum of best segments, etc...
type Run struct {
	id               uuid.UUID
	splitFileVersion int
	_                time.Duration
	_                bool
	_                []SplitPayload
}

func (r *Run) GetPayload() RunPayload {
	return RunPayload{
		ID:               r.id,
		SplitFileVersion: r.splitFileVersion,
	}
}

func NewRunFromPayload(payload RunPayload) Run {
	return Run{
		id:               payload.ID,
		splitFileVersion: payload.SplitFileVersion,
	}
}
