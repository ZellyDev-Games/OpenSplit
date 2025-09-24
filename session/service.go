package session

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zellydev-games/opensplit/logger"
)

const splitDebounce = 120 * time.Millisecond

type SplitResult int

const (
	SplitNoop SplitResult = iota
	SplitStarted
	SplitAdvanced
	SplitFinished
	SplitReset
)

type State byte

const (
	Idle State = iota
	Running
	Paused
	Finished
)

// Split represents an advancement of a run through the Segments.
//
// Split identifies a completed segment, and how long that segment took
type Split struct {
	SplitIndex        int
	SplitSegmentID    uuid.UUID
	CurrentCumulative time.Duration
	CurrentDuration   time.Duration
}

// Segment represents a portion of a game that you want to time (e.g. "Level 1")
type Segment struct {
	ID      uuid.UUID
	Name    string
	Gold    time.Duration
	Average time.Duration
	PB      time.Duration
}

// Run is a snapshot of a SplitFile along with additional data to track a run
type Run struct {
	ID               uuid.UUID
	TotalTime        time.Duration
	Splits           []Split
	Completed        bool
	SplitFileVersion int
}

type SplitFile struct {
	ID           uuid.UUID
	GameName     string
	GameCategory string
	Version      int
	Attempts     int
	Segments     []Segment
	SOB          time.Duration
	Runs         []Run
	PB           *Run
	PBTotalTime  time.Duration
}

// Service represents the current state of a run.
//
// It is the primary glue that brings together a Timer, SplitFile, Run history, tracks the status of the
// current Run / SplitFile, and communicates timer updates to the frontend
// If there's one struct that's key to understand in OpenSplit, it's this one.
type Service struct {
	mu                  sync.Mutex
	timer               Timer
	loadedSplitFile     *SplitFile
	currentRun          *Run
	currentSegmentIndex int
	sessionState        State
	lastSplitTime       time.Time
	repository          Repository
	dirty               bool
}

// NewService creates a new Service from the passed in components.
//
// Generally in real code splitFile should be nil and will be populated by the
// statemachine.Service via UpdateSplitFile or LoadSplitFile
// Timer updates will be sent over the timeUpdatedChannel at approximately 60FPS.
func NewService(repo Repository, timer Timer) *Service {
	service := &Service{
		timer:               timer,
		currentSegmentIndex: -1,
		repository:          repo,
	}
	return service
}

// Split starts, advances, finishes, or resets a run depending on the state
func (s *Service) Split() SplitResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t := time.Now(); s.lastSplitTime.Add(splitDebounce).After(t) {
		return SplitNoop
	} else {
		s.lastSplitTime = t
	}

	switch s.sessionState {
	case Idle:
		// Start a new run
		if s.loadedSplitFile == nil {
			logger.Debug("Split() called with no loaded splitfile.  NO-OP")
			return SplitNoop
		}

		if len(s.loadedSplitFile.Segments) == 0 {
			logger.Debug("Split() called on run with no Segments, NO-OP")
			return SplitNoop
		}
		s.resetLocked()
		s.timer.Start()
		s.loadedSplitFile.Attempts++
		s.sessionState = Running
		s.currentSegmentIndex = 0
		s.currentRun = &Run{
			ID:               uuid.New(),
			SplitFileVersion: s.loadedSplitFile.Version,
		}
		s.dirty = true
		return SplitStarted
	case Running:
		// defensive, prevents us from panic in case something really went wrong
		if s.currentSegmentIndex < 0 || s.currentSegmentIndex >= len(s.loadedSplitFile.Segments) {
			logger.Warn(
				fmt.Sprintf("Split() called in Running state, but current segment index is out of bounds: %d",
					s.currentSegmentIndex))
			return SplitNoop
		}
		now := s.timer.GetCurrentTime()
		prev := time.Duration(0)
		if s.currentSegmentIndex > 0 {
			prev = s.currentRun.Splits[s.currentSegmentIndex-1].CurrentCumulative
		}
		segTime := now - prev

		s.currentRun.Splits = append(s.currentRun.Splits, Split{
			SplitIndex:        s.currentSegmentIndex,
			SplitSegmentID:    s.loadedSplitFile.Segments[s.currentSegmentIndex].ID,
			CurrentCumulative: now,
			CurrentDuration:   segTime,
		})
		s.dirty = true

		if s.currentSegmentIndex >= len(s.loadedSplitFile.Segments)-1 {
			s.timer.Pause()
			s.sessionState = Finished
			s.currentRun.TotalTime = now
			return SplitFinished
		}
		s.currentSegmentIndex++
		return SplitAdvanced
	case Finished:
		s.resetLocked()
		return SplitReset
	case Paused:
		logger.Debug("Split() called with paused - NO-OP")
		return SplitNoop
	}

	return SplitNoop
}

// Undo cancels a Split()
func (s *Service) Undo() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentRun == nil || s.currentSegmentIndex <= 0 || s.sessionState == Idle || len(s.currentRun.Splits) == 0 {
		return
	}

	s.currentRun.Splits = s.currentRun.Splits[:len(s.currentRun.Splits)-1]
	if s.sessionState == Finished {
		s.sessionState = Running
	}

	s.currentSegmentIndex = len(s.currentRun.Splits)

	if s.currentSegmentIndex != 0 {
		s.currentRun.TotalTime = s.currentRun.Splits[len(s.currentRun.Splits)-1].CurrentCumulative
	} else {
		s.currentRun.TotalTime = 0
	}
	s.dirty = true
}

// Skip sets the current segment to the next one without recording a split
func (s *Service) Skip() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentRun == nil ||
		s.currentSegmentIndex >= len(s.loadedSplitFile.Segments)-1 ||
		s.sessionState == Idle ||
		s.sessionState == Finished {
		return
	}
	s.currentSegmentIndex++
	s.dirty = true
}

// Pause toggles the pause state of a run
func (s *Service) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sessionState != Running && s.sessionState != Paused {
		return
	}
	if s.sessionState == Running {
		s.sessionState = Paused
		s.timer.Pause()
	} else {
		s.sessionState = Running
		s.timer.Start()
	}
}

// Load loads a split file from the configured Repository and sets it as the current Run
func (s *Service) Load() error {
	sf, err := s.repository.Load()
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.loadedSplitFile = sf
	s.resetLocked()
	s.dirty = false
	s.mu.Unlock()
	return nil
}

// Save persists the current Run with the configured Repository
func (s *Service) Save() error {
	if s.currentRun == nil {
		return nil
	}
	s.mu.Lock()
	file := s.loadedSplitFile
	s.mu.Unlock()
	err := s.repository.Save(file)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.dirty = false
	s.mu.Unlock()
	return nil
}

// SaveAs saves the current Run with the configured Repository, intending to force the file dialog path
func (s *Service) SaveAs() error {
	if s.currentRun == nil {
		return nil
	}
	s.mu.Lock()
	sf := s.loadedSplitFile
	s.mu.Unlock()

	err := s.repository.SaveAs(sf)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.dirty = false
	s.mu.Unlock()
	return nil
}

// Reset stops any current run and brings the system back to a default state.
func (s *Service) Reset() {
	s.mu.Lock()
	s.resetLocked()
	s.mu.Unlock()
}

// CloseRun unloads the loaded Run, and resets the system.
func (s *Service) CloseRun() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.currentRun = nil
	s.resetLocked()
}

func (s *Service) SplitFile() *SplitFile {
	return s.loadedSplitFile
}

// Dirty returns the unsaved changes status of the session
func (s *Service) Dirty() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.dirty }

// State returns the session State
func (s *Service) State() State { s.mu.Lock(); defer s.mu.Unlock(); return s.sessionState }

// Index returns the current segment index of the session
func (s *Service) Index() int { s.mu.Lock(); defer s.mu.Unlock(); return s.currentSegmentIndex }

// Run returns the currently loaded Run
//
// Returns true as second param only if result is valid
func (s *Service) Run() (Run, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentRun == nil {
		return Run{}, false
	}
	r := *s.currentRun
	r.Splits = append([]Split(nil), r.Splits...)
	return r, true
}

// resetLocked assumed that the system is under lock when called.
func (s *Service) resetLocked() {
	s.timer.Pause()
	s.timer.Reset()
	if s.currentRun != nil {
		s.loadedSplitFile.Runs = append(s.loadedSplitFile.Runs, *s.currentRun)
		s.loadedSplitFile.BuildStats()
		if s.sessionState == Finished {
			s.currentRun.Completed = true
		}
		s.currentRun = nil
	}

	s.sessionState = Idle
	s.currentSegmentIndex = -1
}
