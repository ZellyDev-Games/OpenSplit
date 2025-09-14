package session

import (
	"OpenSplit/logger"
	"OpenSplit/utils"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SegmentPayload is a snapshot of a given Segment useful for communicating state while protecting internal data.
type SegmentPayload struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	BestTime string `json:"best_time"`
	Average  string `json:"average_time"`
}

// Segment represents a portion of a game that you want to time (e.g. "Level 1")
type Segment struct {
	id          uuid.UUID
	name        string
	bestTime    time.Duration
	averageTime time.Duration
}

// NewFromPayload creates a new Segment from the given SegmentPayload.
//
// This is a pattern used often in OpenSplit where you fetch a payload, modify it, then pass it into a modification or
// creation func to persist changes internally.
func NewFromPayload(payload SegmentPayload) (Segment, error) {
	if payload.ID == "" {
		payload.ID = uuid.New().String()
	}
	bestTime, err := utils.ParseStringToTime(payload.BestTime)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to parse best time: %s", err))
	}

	averageTime, err := utils.ParseStringToTime(payload.Average)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to parse average time: %s", err))
	}

	return Segment{id: uuid.MustParse(payload.ID), name: payload.Name, bestTime: bestTime, averageTime: averageTime}, err
}

// GetPayload retrieves a SegmentPayload representing the state of the Segment
func (s *Segment) GetPayload() SegmentPayload {
	return SegmentPayload{
		ID:       s.id.String(),
		Name:     s.name,
		BestTime: utils.FormatTimeToString(s.bestTime),
		Average:  utils.FormatTimeToString(s.averageTime),
	}
}
