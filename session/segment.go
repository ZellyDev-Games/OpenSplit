package session

import (
	"time"

	"github.com/google/uuid"
)

type SegmentPayload struct {
	ID       uuid.UUID     `json:"id"`
	Name     string        `json:"name"`
	BestTime time.Duration `json:"best_time"`
	Average  time.Duration `json:"average_time"`
}

type Segment struct {
	id          uuid.UUID
	name        string
	bestTime    time.Duration
	averageTime time.Duration
}

func NewSegment(id uuid.UUID, name string) *Segment {
	return &Segment{id: id, name: name}
}

func (s *Segment) GetPayload() SegmentPayload {
	return SegmentPayload{
		ID:       s.id,
		Name:     s.name,
		BestTime: s.bestTime,
		Average:  s.averageTime,
	}
}
