package session

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/zellydev-games/opensplit/logger"
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
	Dirty        bool             `json:"dirty"`
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
	dirty        bool
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
			Formatted: FormatTimeToString(s.sob),
		},
		WindowParams: s.windowParams,
		Dirty:        s.dirty,
	}
}

func SplitFileChanged(file1 SplitFilePayload, file2 SplitFilePayload) bool {
	return !reflect.DeepEqual(file1.Segments, file2.Segments) || !reflect.DeepEqual(file1.GameCategory, file2.GameCategory)
}

// Save uses the configured Persister to save the SplitFile to the configured storage
//
// Use Save instead of UpdateSplitFile when you want to save new runs or BuildStats without changes to data
// (e.g. NOT changing the Game Name, Category, or segments).
// This function will never bump the split file version.
func (s *SplitFile) Save(persister Persister, windowParams WindowParams) error {
	s.windowParams.Width = windowParams.Width
	s.windowParams.Height = windowParams.Height
	s.windowParams.X = windowParams.X
	s.windowParams.Y = windowParams.Y

	err := persister.Save(s.GetPayload())
	if err != nil {
		var cancelled = &UserCancelledSave{}
		if errors.As(err, cancelled) {
			logger.Debug("user cancelled save")
			logger.Error(fmt.Sprintf("failed to save split file with Save: %s", err))
		}
	}
	return err
}

// UpdateSplitFile uses the configured Persister to save the SplitFile to the configured storage.
//
// It creates a SplitFile from the given SplitFilePayload, then returns a pointer to it.
func UpdateSplitFile(runtimeProvider RuntimeProvider, persister Persister, existingSplitFile *SplitFilePayload, payload SplitFilePayload) (*SplitFile, error) {
	newSplitFile, err := newFromPayload(payload)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to parse split file payload: %s", err))
		return nil, err
	}

	if existingSplitFile == nil {
		// this is a brand new splitfile
		newSplitFile.version = 1
	} else {
		// Check if the segments or the category changed.  If so bump the version.
		if SplitFileChanged(*existingSplitFile, payload) {
			if runtimeProvider != nil {
				res, err := runtimeProvider.MessageDialog(runtime.MessageDialogOptions{
					Type:    runtime.QuestionDialog,
					Title:   "Gold Reset",
					Message: "Changing segments will reset golds and averages for this split file. Proceed?",
				})
				if err != nil {
					return nil, err
				}
				if res != "yes" {
					return nil, UserCancelledSave{err}
				}
			}
			newSplitFile.version = existingSplitFile.Version + 1
		}
	}

	err = persister.Save(newSplitFile.GetPayload())
	if err != nil {
		var cancelled = &UserCancelledSave{}
		if errors.As(err, cancelled) {
			logger.Debug("user cancelled save")
			return nil, err
		}
		logger.Error(fmt.Sprintf("failed to save split file: %s", err))
		return nil, err
	}

	logger.Debug("sending session update from update split file")
	return newSplitFile, err
}

// LoadSplitFile retrieves a SplitFilePayload from Persister configured storage.
//
// It creates a new SplitFile from the retrieved SplitFilePayload, sets that as the loaded split file, and resets the
// system.
func LoadSplitFile(persister Persister) (*SplitFile, error) {
	newSplitFilePayload, err := persister.Load()
	if err != nil {
		var userCancelled = &UserCancelledSave{}
		if !errors.As(err, userCancelled) {
			logger.Error(fmt.Sprintf("failed to load split file: %s", err))
		}
		return nil, err
	}

	return newFromPayload(newSplitFilePayload)
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

	// ensure sane window params
	payload.WindowParams.Height = max(200, payload.WindowParams.Height)
	payload.WindowParams.Width = max(200, payload.WindowParams.Width)

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
