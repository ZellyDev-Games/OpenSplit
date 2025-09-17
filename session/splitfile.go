package session

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/zellydev-games/opensplit/logger"
)

// SplitFilePayload is a snapshot of a SplitFile
//
// Used to communicate the state of a SplitFile to the frontend and Persister implementations without exposing internals.
type SplitFilePayload struct {
	ID           string                `json:"id"`
	Version      int                   `json:"version"`
	GameName     string                `json:"game_name"`
	GameCategory string                `json:"game_category"`
	Segments     []SegmentPayload      `json:"segments"`
	Attempts     int                   `json:"attempts"`
	Runs         []RunPayload          `json:"runs"`
	Stats        SplitFileStatsPayload `json:"stats"`
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
		runPayloads = append(runPayloads, run.getPayload())
	}

	stats := s.Stats()
	statsPayload, err := stats.GetPayload()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get Stats payload: %s", err))
	}
	return SplitFilePayload{
		ID:           s.id.String(),
		GameName:     s.gameName,
		GameCategory: s.gameCategory,
		Segments:     segmentPayloads,
		Attempts:     s.attempts,
		Runs:         runPayloads,
		Version:      s.version,
		Stats:        statsPayload,
	}
}

func SplitFileChanged(file1 SplitFilePayload, file2 SplitFilePayload) bool {
	return !reflect.DeepEqual(file1.Segments, file2.Segments) || !reflect.DeepEqual(file1.GameCategory, file2.GameCategory)
}

func newFromPayload(payload SplitFilePayload) (*SplitFile, error) {
	var segments []Segment
	for _, segment := range payload.Segments {
		segments = append(segments, NewFromPayload(segment))
	}

	var runs []Run
	for _, run := range payload.Runs {
		newRun := newRunFromPayload(run)
		runs = append(runs, newRun)
	}

	if payload.ID == uuid.Nil.String() {
		payload.ID = uuid.New().String()
	}

	return &SplitFile{
		id:           uuid.MustParse(payload.ID),
		gameName:     payload.GameName,
		gameCategory: payload.GameCategory,
		attempts:     payload.Attempts,
		segments:     segments,
		runs:         runs,
		version:      payload.Version,
	}, nil
}
