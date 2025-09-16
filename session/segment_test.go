package session

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewFromPayload(t *testing.T) {
	payload := SegmentPayload{
		ID:   "4bc1a05c-d4f3-4095-887f-519e2fbb54f3",
		Name: "Test Segment",
	}

	s := NewFromPayload(payload)
	if s.id != uuid.MustParse("4bc1a05c-d4f3-4095-887f-519e2fbb54f3") ||
		s.name != "Test Segment" {
		t.Errorf("NewFromPayload did not return expected new segment")
	}
}

func TestGetPayload(t *testing.T) {
	s := Segment{
		id:   uuid.MustParse("4bc1a05c-d4f3-4095-887f-519e2fbb54f3"),
		name: "Test Segment",
	}

	payload := s.GetPayload()
	if payload.ID != "4bc1a05c-d4f3-4095-887f-519e2fbb54f3" ||
		payload.Name != "Test Segment" {
		t.Errorf("getPayload did not return expected payload got %v", payload)
	}
}
