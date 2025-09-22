package session

import (
	"time"

	"github.com/google/uuid"
)

type RunPayload struct {
	ID               string         `json:"id"`
	SplitFileVersion int            `json:"splitfile_version"`
	TotalTime        StatTime       `json:"total_time"`
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
	splits           []Split
}

// getPayload returns a snapshot of a Run
//
// getPayload, modify the payload, then send it to newRunFromPayload and persist the result of that to make changes.
func (r *Run) getPayload() RunPayload {
	splitPayloads := make([]SplitPayload, len(r.splits))
	for i, s := range r.splits {
		splitPayloads[i] = s.getPayload()
	}

	return RunPayload{
		ID:               r.id.String(),
		SplitFileVersion: r.splitFileVersion,
		TotalTime: StatTime{
			Raw:       r.totalTime.Milliseconds(),
			Formatted: FormatTimeToString(r.totalTime),
		},
		Completed:     r.completed,
		SplitPayloads: splitPayloads,
	}
}

// newRunFromPayload creates a new Run from a RunPayload.
//
// Useful for making stateful updates to a Run without exposing internal data structure or presentation.
func newRunFromPayload(payload RunPayload) Run {
	var splits []Split
	for _, s := range payload.SplitPayloads {
		splits = append(splits, Split{
			splitIndex:      s.SplitIndex,
			splitSegmentID:  uuid.MustParse(s.SplitSegmentID),
			currentDuration: PayloadRawTimeToDuration(s.CurrentTime.Raw),
		})
	}

	return Run{
		id:               uuid.MustParse(payload.ID),
		splitFileVersion: payload.SplitFileVersion,
		totalTime:        time.Duration(payload.TotalTime.Raw) * time.Millisecond,
		completed:        payload.Completed,
		splits:           splits,
	}
}
