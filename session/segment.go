package session

import "github.com/google/uuid"

type Segment struct {
	id   uuid.UUID
	name string
}

func NewSegment(id uuid.UUID, name string) *Segment {
	return &Segment{id: id, name: name}
}

type SegmentStats struct {
}
