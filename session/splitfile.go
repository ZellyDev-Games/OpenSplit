package session

import (
	"reflect"

	"github.com/google/uuid"
)

// SplitFilePayload is a snapshot of a SplitFile
//
// Used to communicate the state of a SplitFile to the frontend and Persister implementations without exposing internals.
type SplitFilePayload struct {
	ID           uuid.UUID        `json:"id"`
	Version      int              `json:"version"`
	GameName     string           `json:"game_name"`
	GameCategory string           `json:"game_category"`
	Segments     []SegmentPayload `json:"segments"`
	Attempts     int              `json:"attempts"`
	Runs         []RunPayload     `json:"runs"`
}

// SplitFile represents the data and history of a game/category combo.
type SplitFile struct {
	id           uuid.UUID
	version      int
	gameName     string
	gameCategory string
	segments     []Segment
	attempts     int
	runs         []Run
}

// NewSplitFile constructor for SplitFile
func NewSplitFile(gameName string, gameCategory string, segments []Segment, attempts int, runs []Run) *SplitFile {
	return &SplitFile{
		gameName:     gameName,
		gameCategory: gameCategory,
		segments:     segments,
		attempts:     attempts,
		runs:         runs,
	}
}

// NewAttempt provides a public function to increment the attempts count
func (s *SplitFile) NewAttempt() {
	s.attempts++
}

// SetAttempts provides a public function to set the attempts count
func (s *SplitFile) SetAttempts(attempts int) {
	s.attempts = attempts
}

// GetPayload gets a snapshot of the SplitFile.  Useful for communicating the state of the file while protecting the
// internal data.
func (s *SplitFile) GetPayload() SplitFilePayload {
	var segmentPayloads []SegmentPayload
	for _, segment := range s.segments {
		segmentPayloads = append(segmentPayloads, segment.GetPayload())
	}

	var runPayloads []RunPayload
	for _, run := range s.runs {
		runPayloads = append(runPayloads, run.GetPayload())
	}

	return SplitFilePayload{
		ID:           s.id,
		GameName:     s.gameName,
		GameCategory: s.gameCategory,
		Segments:     segmentPayloads,
		Attempts:     s.attempts,
		Runs:         runPayloads,
		Version:      s.version,
	}
}

func SplitFileChanged(file1 SplitFilePayload, file2 SplitFilePayload) bool {
	// Set fields that we don't want to cause a version change to be equal
	// Note: This is a copy of the payloads, so mutation here is safe for the caller.
	file1.Runs = nil
	file2.Runs = nil
	file1.Attempts = 0
	file2.Attempts = 0
	file1.GameName = ""
	file2.GameName = ""
	return !reflect.DeepEqual(file1, file2)
}

func newFromPayload(payload SplitFilePayload) (*SplitFile, error) {
	var segments []Segment
	for _, segment := range payload.Segments {
		newSegment, err := NewFromPayload(segment)
		if err != nil {
			return nil, err
		}
		segments = append(segments, newSegment)
	}

	var runs []Run
	for _, run := range payload.Runs {
		newRun := NewRunFromPayload(run)
		runs = append(runs, newRun)
	}

	var emptyUUID = uuid.UUID{}
	if payload.ID == emptyUUID {
		payload.ID = uuid.New()
	}

	return &SplitFile{
		id:           payload.ID,
		gameName:     payload.GameName,
		gameCategory: payload.GameCategory,
		attempts:     payload.Attempts,
		segments:     segments,
		runs:         runs,
		version:      payload.Version,
	}, nil
}
