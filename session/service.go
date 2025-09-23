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

// Repository defines a contract for a persistence provider to operate against
type Repository interface {
	Load() (*Run, error)
	Save(*Run) error
	SaveAs(run *Run) error
}

// Split represents an advancement of a run through the segments.
//
// Split identifies a completed segment, and how long that segment took
type Split struct {
	splitIndex        int
	splitSegmentID    uuid.UUID
	currentCumulative time.Duration
	currentDuration   time.Duration
}

// Segment represents a portion of a game that you want to time (e.g. "Level 1")
type Segment struct {
	id      uuid.UUID
	name    string
	gold    time.Duration
	average time.Duration
	pb      time.Duration
}

// Run is a snapshot of a SplitFile along with additional data to track a run
type Run struct {
	id               uuid.UUID
	splitFileID      uuid.UUID
	splitFileVersion int
	gameName         string
	gameCategory     string
	segments         []Segment
	attempts         int
	sob              StatTime
	totalTime        time.Duration
	splits           []Split
}

// Service represents the current state of a run.
//
// It is the primary glue that brings together a Timer, SplitFile, Run history, tracks the status of the
// current Run / SplitFile, and communicates timer updates to the frontend
// If there's one struct that's key to understand in OpenSplit, it's this one.
type Service struct {
	mu                  sync.Mutex
	timer               Timer
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
		if s.currentRun == nil {
			logger.Debug("Split() called with no loaded Run.  NO-OP")
			return SplitNoop
		}

		if len(s.currentRun.segments) == 0 {
			logger.Debug("Split() called on run with no segments, NO-OP")
			return SplitNoop
		}
		s.resetLocked()
		s.timer.Start()
		s.currentRun.attempts++
		s.sessionState = Running
		s.currentSegmentIndex = 0
		s.dirty = true
		return SplitStarted
	case Running:
		// defensive, prevents us from panic in case something really went wrong
		if s.currentSegmentIndex < 0 || s.currentSegmentIndex >= len(s.currentRun.segments) {
			logger.Warn(
				fmt.Sprintf("Split() called in Running state, but current segment index is out of bounds: %d",
					s.currentSegmentIndex))
			return SplitNoop
		}
		now := s.timer.GetCurrentTime()
		prev := time.Duration(0)
		if s.currentSegmentIndex > 0 {
			prev = s.currentRun.splits[s.currentSegmentIndex-1].currentCumulative
		}
		segTime := now - prev

		s.currentRun.splits = append(s.currentRun.splits, Split{
			splitIndex:        s.currentSegmentIndex,
			splitSegmentID:    s.currentRun.segments[s.currentSegmentIndex].id,
			currentCumulative: now,
			currentDuration:   segTime,
		})
		s.dirty = true

		if s.currentSegmentIndex >= len(s.currentRun.segments)-1 {
			s.timer.Pause()
			s.sessionState = Finished
			s.currentRun.totalTime = now
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

	if s.currentRun == nil || s.currentSegmentIndex <= 0 || s.sessionState == Idle || len(s.currentRun.splits) == 0 {
		return
	}

	s.currentRun.splits = s.currentRun.splits[:len(s.currentRun.splits)-1]
	if s.sessionState == Finished {
		s.sessionState = Running
	}

	s.currentSegmentIndex = len(s.currentRun.splits)

	if s.currentSegmentIndex != 0 {
		s.currentRun.totalTime = s.currentRun.splits[len(s.currentRun.splits)-1].currentCumulative
	} else {
		s.currentRun.totalTime = 0
	}
	s.dirty = true
}

// Skip sets the current segment to the next one without recording a split
func (s *Service) Skip() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentRun == nil ||
		s.currentSegmentIndex >= len(s.currentRun.segments)-1 ||
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
	run, err := s.repository.Load()
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.currentRun = run
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
	run := s.currentRun
	s.mu.Unlock()
	err := s.repository.Save(run)
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
	run := s.currentRun
	s.mu.Unlock()

	err := s.repository.SaveAs(run)
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
	r.segments = append([]Segment(nil), r.segments...)
	r.splits = append([]Split(nil), r.splits...)
	return r, true
}

// resetLocked assumed that the system is under lock when called.
func (s *Service) resetLocked() {
	s.timer.Pause()
	s.timer.Reset()
	if s.currentRun != nil {
		s.currentRun.totalTime = 0
		clear(s.currentRun.splits)
		s.currentRun.splits = s.currentRun.splits[:0]
	}

	s.sessionState = Idle
	s.currentSegmentIndex = -1
}
