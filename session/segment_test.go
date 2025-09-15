package session

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewFromPayload(t *testing.T) {
	payload := SegmentPayload{
		ID:       "4bc1a05c-d4f3-4095-887f-519e2fbb54f3",
		Name:     "Test Segment",
		BestTime: "00:00:01.00",
		Average:  "00:00:02.00",
	}

	s, _ := NewFromPayload(payload)
	if s.id != uuid.MustParse("4bc1a05c-d4f3-4095-887f-519e2fbb54f3") ||
		s.name != "Test Segment" ||
		s.bestTime != time.Second*1 ||
		s.averageTime != time.Second*2 {
		t.Errorf("NewFromPayload did not return expected new segment")
	}
}

func TestGetPayload(t *testing.T) {
	s := Segment{
		id:          uuid.MustParse("4bc1a05c-d4f3-4095-887f-519e2fbb54f3"),
		name:        "Test Segment",
		bestTime:    time.Second * 1,
		averageTime: time.Second * 2,
	}

	payload := s.GetPayload()
	if payload.ID != "4bc1a05c-d4f3-4095-887f-519e2fbb54f3" ||
		payload.Name != "Test Segment" ||
		payload.Average != "00:00:02.00" ||
		payload.BestTime != "00:00:01.00" {
		t.Errorf("GetPayload did not return expected payload got %v", payload)
	}
}
