package session

import (
	"context"
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

// Timer is an interface that a stopwatch service must implement to be used by session.Service
type Timer interface {
	Startup(context.Context)
	IsRunning() bool
	Run()
	Start()
	Pause()
	Reset()
	GetCurrentTime() time.Duration
}

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
	ID       uuid.UUID
	Name     string
	Gold     time.Duration
	Average  time.Duration
	PB       time.Duration
	Children []Segment
}

// Run is a snapshot of a SplitFile along with additional data to track a run
type Run struct {
	ID               uuid.UUID
	TotalTime        time.Duration
	Splits           []*Split
	Segments         []Segment
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
	WindowX      int
	WindowY      int
	WindowHeight int
	WindowWidth  int
	SOB          time.Duration
	Runs         []Run
	PB           *Run
}

// Service represents the current state of a run.
//
// It is the primary glue that brings together a Timer, SplitFile, Run history, tracks the status of the
// current Run / SplitFile, and communicates timer updates to the frontend
// If there's one struct that's key to understand in OpenSplit, it's this one.
type Service struct {
	mu                   sync.Mutex
	timer                Timer
	loadedSplitFile      *SplitFile
	runtimeSegments      []Segment
	currentRun           *Run
	currentSegmentIndex  int
	sessionState         State
	lastSplitTime        time.Time
	dirty                bool
	sessionUpdateChannel chan *Service
}

// NewService creates a new Service from the passed in components.
//
// Generally in real code splitFile should be nil and will be populated by the
// statemachine.Service via UpdateSplitFile or LoadSplitFile
// Timer updates will be sent over the timeUpdatedChannel at approximately 60FPS.
func NewService(timer Timer) (*Service, chan *Service) {
	service := &Service{
		timer:                timer,
		currentSegmentIndex:  -1,
		sessionUpdateChannel: make(chan *Service, 128),
	}

	return service, service.sessionUpdateChannel
}

func (s *Service) SetLoadedSplitFile(sf *SplitFile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.sendUpdate()

	s.loadedSplitFile = sf
	s.runtimeSegments = FlattenSegments(sf.Segments)

	s.currentRun = nil
	s.currentSegmentIndex = -1
	s.sessionState = Idle
	s.dirty = false
}

// Split starts, advances, finishes, or resets a run depending on the state
func (s *Service) Split() SplitResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.sendUpdate()

	if !s.debounced() {
		return SplitNoop
	}

	switch s.sessionState {
	case Idle:
		return s.startNewRun()
	case Running:
		return s.advanceRun()
	case Finished:
		s.resetLocked()
		return SplitReset
	case Paused:
		return SplitNoop
	}
	return SplitNoop
}

// Undo cancels a Split()
func (s *Service) Undo() {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.sendUpdate()

	if s.currentRun == nil || s.currentSegmentIndex <= 0 || s.sessionState == Idle {
		return
	}

	s.currentSegmentIndex--

	// nil out the split at the current index
	s.currentRun.Splits[s.currentSegmentIndex] = nil

	// recompute TotalTime from last non-nil split
	total := time.Duration(0)
	for i := s.currentSegmentIndex - 1; i >= 0; i-- {
		if sp := s.currentRun.Splits[i]; sp != nil {
			total = sp.CurrentCumulative
			break
		}
	}
	s.currentRun.TotalTime = total

	if s.sessionState == Finished {
		s.sessionState = Running
		s.timer.Start()
	}

	s.dirty = true
}

// Skip sets the current segment to the next one without recording a split
func (s *Service) Skip() {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.sendUpdate()

	if s.currentRun == nil ||
		s.currentSegmentIndex >= len(s.runtimeSegments)-1 ||
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
	defer s.sendUpdate()

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

// Reset stops any current run and brings the system back to a default state.
func (s *Service) Reset() {
	s.mu.Lock()
	s.resetLocked()
	s.mu.Unlock()
	s.sendUpdate()
}

// CloseRun unloads the loaded Run, and resets the system.
func (s *Service) CloseRun() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.currentRun = nil
	s.resetLocked()
}

func (s *Service) SplitFile() *SplitFile {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deepCopySplitFile()
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
	r.Splits = append([]*Split(nil), r.Splits...)
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

func (s *Service) debounced() bool {
	if t := time.Now(); s.lastSplitTime.Add(splitDebounce).After(t) {
		return false
	} else {
		s.lastSplitTime = t
		return true
	}
}

func (s *Service) startNewRun() SplitResult {
	// Start a new run
	if s.loadedSplitFile == nil {
		logger.Debug("Split() called with no loaded dto.  NO-OP")
		return SplitNoop
	}

	if len(s.runtimeSegments) == 0 {
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
		Splits:           make([]*Split, len(s.runtimeSegments)),
		Segments:         append([]Segment(nil), s.runtimeSegments...),
		SplitFileVersion: s.loadedSplitFile.Version,
	}
	s.dirty = true
	return SplitStarted
}

func (s *Service) advanceRun() SplitResult {
	if s.currentSegmentIndex < 0 || s.currentSegmentIndex >= len(s.runtimeSegments) {
		logger.Warn(
			fmt.Sprintf("Split() called in Running state, but current segment index is out of bounds: %d",
				s.currentSegmentIndex))
		return SplitNoop
	}
	now := s.timer.GetCurrentTime()

	// find prev cumulative from the last non-nil split
	prev := time.Duration(0)
	for i := s.currentSegmentIndex - 1; i >= 0; i-- {
		if sp := s.currentRun.Splits[i]; sp != nil {
			prev = sp.CurrentCumulative
			break
		}
	}

	segTime := now - prev

	s.currentRun.Splits[s.currentSegmentIndex] = &Split{
		SplitIndex:        s.currentSegmentIndex,
		SplitSegmentID:    s.runtimeSegments[s.currentSegmentIndex].ID,
		CurrentCumulative: now,
		CurrentDuration:   segTime,
	}

	s.dirty = true
	s.currentSegmentIndex++

	if s.currentSegmentIndex > len(s.runtimeSegments)-1 {
		s.timer.Pause()
		s.sessionState = Finished
		s.currentRun.TotalTime = now
		return SplitFinished
	}
	return SplitAdvanced
}

// sendUpdate must be called when s.mu is held by the caller
func (s *Service) sendUpdate() {
	if s.sessionUpdateChannel == nil {
		return
	}
	select {
	case s.sessionUpdateChannel <- s:
	default:
	}
}

func (s *Service) deepCopySplitFile() *SplitFile {
	if s.loadedSplitFile == nil {
		return nil
	}

	segments := deepCopySegments(s.loadedSplitFile.Segments)

	runs := make([]Run, 0, len(s.loadedSplitFile.Runs))
	for _, run := range s.loadedSplitFile.Runs {
		flatSegs := make([]Segment, len(run.Segments))
		for i, segment := range run.Segments {
			flatSegs[i] = Segment{
				ID:      segment.ID,
				Name:    segment.Name,
				Gold:    segment.Gold,
				Average: segment.Average,
				PB:      segment.PB,
			}
		}

		// Copy splits
		splits := make([]*Split, len(run.Segments))
		for _, split := range run.Splits {
			if split == nil {
				continue
			}
			splits[split.SplitIndex] = &Split{
				SplitIndex:        split.SplitIndex,
				SplitSegmentID:    split.SplitSegmentID,
				CurrentCumulative: split.CurrentCumulative,
				CurrentDuration:   split.CurrentDuration,
			}
		}

		runs = append(runs, Run{
			ID:               run.ID,
			SplitFileVersion: run.SplitFileVersion,
			TotalTime:        run.TotalTime,
			Splits:           splits,
			Completed:        run.Completed,
			Segments:         flatSegs,
		})
	}

	var pbRun *Run
	if s.loadedSplitFile.PB != nil {
		src := s.loadedSplitFile.PB

		flatSegs := make([]Segment, len(src.Segments))
		for i, segment := range src.Segments {
			flatSegs[i] = Segment{
				ID:      segment.ID,
				Name:    segment.Name,
				Gold:    segment.Gold,
				Average: segment.Average,
				PB:      segment.PB,
			}
		}

		splits := make([]*Split, len(src.Segments))
		for _, split := range src.Splits {
			if split == nil {
				continue
			}
			splits[split.SplitIndex] = &Split{
				SplitIndex:        split.SplitIndex,
				SplitSegmentID:    split.SplitSegmentID,
				CurrentCumulative: split.CurrentCumulative,
				CurrentDuration:   split.CurrentDuration,
			}
		}

		pbRun = &Run{
			ID:               src.ID,
			SplitFileVersion: src.SplitFileVersion,
			TotalTime:        src.TotalTime,
			Splits:           splits,
			Completed:        src.Completed,
			Segments:         flatSegs,
		}
	}

	return &SplitFile{
		ID:           s.loadedSplitFile.ID,
		GameName:     s.loadedSplitFile.GameName,
		GameCategory: s.loadedSplitFile.GameCategory,
		Version:      s.loadedSplitFile.Version,
		Attempts:     s.loadedSplitFile.Attempts,
		Segments:     segments,
		WindowX:      s.loadedSplitFile.WindowX,
		WindowY:      s.loadedSplitFile.WindowY,
		WindowHeight: s.loadedSplitFile.WindowHeight,
		WindowWidth:  s.loadedSplitFile.WindowWidth,
		SOB:          s.loadedSplitFile.SOB,
		Runs:         runs,
		PB:           pbRun,
	}
}

func FlattenSegments(list []Segment) []Segment {
	var out []Segment
	for _, s := range list {

		// Parent segment WITHOUT children
		out = append(out, Segment{
			ID:      s.ID,
			Name:    s.Name,
			Gold:    s.Gold,
			Average: s.Average,
			PB:      s.PB,
			// no children
		})

		// Recursively append children
		if len(s.Children) > 0 {
			out = append(out, FlattenSegments(s.Children)...)
		}
	}
	return out
}

func deepCopySegments(list []Segment) []Segment {
	out := make([]Segment, len(list))
	for i, s := range list {
		out[i] = Segment{
			ID:       s.ID,
			Name:     s.Name,
			Gold:     s.Gold,
			Average:  s.Average,
			PB:       s.PB,
			Children: deepCopySegments(s.Children),
		}
	}
	return out
}
