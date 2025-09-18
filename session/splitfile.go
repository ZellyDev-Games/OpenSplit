package session

import (
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/zellydev-games/opensplit/utils"
)

// WindowParams stores the last size and position the user set splitter window while this file was loaded
type WindowParams struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

// NewDefaultWindowParams returns a WindowParams with sensible defaults
func NewDefaultWindowParams() WindowParams {
	return WindowParams{
		Width:  350,
		Height: 530,
		X:      200,
		Y:      200,
	}
}

// SplitFilePayload is a snapshot of a SplitFile
//
// Used to communicate the state of a SplitFile to the frontend and Persister implementations without exposing internals.
type SplitFilePayload struct {
	ID           string           `json:"id"`
	Version      int              `json:"version"`
	GameName     string           `json:"game_name"`
	GameCategory string           `json:"game_category"`
	Segments     []SegmentPayload `json:"segments"`
	Attempts     int              `json:"attempts"`
	Runs         []RunPayload     `json:"runs"`
	SOB          StatTime         `json:"SOB"`
	WindowParams WindowParams     `json:"window_params"`
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
	sob          time.Duration
	windowParams WindowParams
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
	var segmentPayloads = make([]SegmentPayload, len(s.segments))
	for i, segment := range s.segments {
		segmentPayloads[i] = segment.GetPayload()
	}

	var runPayloads = make([]RunPayload, len(s.runs))
	for i, run := range s.runs {
		runPayloads[i] = run.getPayload()
	}

	return SplitFilePayload{
		ID:           s.id.String(),
		GameName:     s.gameName,
		GameCategory: s.gameCategory,
		Segments:     segmentPayloads,
		Attempts:     s.attempts,
		Runs:         runPayloads,
		Version:      s.version,
		SOB: StatTime{
			Raw:       s.sob.Milliseconds(),
			Formatted: utils.FormatTimeToString(s.sob),
		},
		WindowParams: s.windowParams,
	}
}

func SplitFileChanged(file1 SplitFilePayload, file2 SplitFilePayload) bool {
	return !reflect.DeepEqual(file1.Segments, file2.Segments) || !reflect.DeepEqual(file1.GameCategory, file2.GameCategory)
}

func newFromPayload(payload SplitFilePayload) (*SplitFile, error) {
	var segments = make([]Segment, len(payload.Segments))
	for i, segment := range payload.Segments {
		segments[i] = NewFromPayload(segment)
	}

	var runs = make([]Run, len(payload.Runs))
	for i, run := range payload.Runs {
		runs[i] = newRunFromPayload(run)
	}

	if payload.ID == uuid.Nil.String() || payload.ID == "" {
		payload.ID = uuid.New().String()
	}

	sf := SplitFile{
		id:           uuid.MustParse(payload.ID),
		gameName:     payload.GameName,
		gameCategory: payload.GameCategory,
		attempts:     payload.Attempts,
		segments:     segments,
		runs:         runs,
		version:      payload.Version,
		windowParams: payload.WindowParams,
	}

	sf.BuildStats()
	return &sf, nil
}
