package session

import (
	"testing"

	"github.com/google/uuid"
)

func TestAttempts(t *testing.T) {
	s := SplitFile{
		gameName:     "Test Game",
		gameCategory: "Test Category",
		segments: []Segment{{
			id:          uuid.UUID{},
			name:        "Test Segment 1",
			bestTime:    1,
			averageTime: 2,
		}},
		attempts: 0,
	}

	s.NewAttempt()
	if s.attempts != 1 {
		t.Errorf("Test failed. Expected 1 attempt, got %d", s.attempts)
	}

	s.SetAttempts(50)
	if s.attempts != 50 {
		t.Errorf("Test failed. Expected 50 attempts, got %d", s.attempts)
	}
}
