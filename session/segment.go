package session

import (
	"OpenSplit/logger"
	"OpenSplit/utils"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SegmentPayload struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	BestTime string `json:"best_time"`
	Average  string `json:"average_time"`
}

type Segment struct {
	id          uuid.UUID
	name        string
	bestTime    time.Duration
	averageTime time.Duration
}

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

func (s *Segment) GetPayload() SegmentPayload {
	return SegmentPayload{
		ID:       s.id.String(),
		Name:     s.name,
		BestTime: utils.FormatTimeToString(s.bestTime),
		Average:  utils.FormatTimeToString(s.averageTime),
	}
}
