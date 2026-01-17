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
	SubtractTime(duration time.Duration)
}

// Split represents an advancement of a run through the DeepCopyLeafSegments.
//
// Split identifies a completed segment, and how long that segment took
type Split struct {
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
	Splits           map[uuid.UUID]Split // uuid key here is a segment ID
	LeafSegments     []Segment
	Completed        bool
	SplitFileVersion int
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
	leafSegments         []*Segment
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

// UpdateWindowDimensions sets the loadedSplitFile window dimensions
// This is needed because there's a subsystem that asynchronously
// updates the splitfile's window dimension information on disk, and the loaded
// session in memory has no idea about those changes, so if the session is copied and saved,
// it clobbers the changes that subsystem made previously.
func (s *Service) UpdateWindowDimensions(x, y, w, h int) {
	s.loadedSplitFile.WindowX = x
	s.loadedSplitFile.WindowY = y
	s.loadedSplitFile.WindowWidth = w
	s.loadedSplitFile.WindowHeight = h
}

func (s *Service) SetLoadedSplitFile(sf SplitFile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.sendUpdate()

	s.loadedSplitFile = &sf
	s.leafSegments = getLeafSegments(sf.Segments, nil)

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
	if s.currentSegmentIndex < 0 || s.currentSegmentIndex >= len(s.leafSegments) {
		logger.Error(
			fmt.Sprintf("Undo() set currentSegmentIndex outside of bounds %d", s.currentSegmentIndex))
		return
	}

	// delete the split at the current index
	segmentID := s.currentRun.LeafSegments[s.currentSegmentIndex].ID
	delete(s.currentRun.Splits, segmentID)

	// recompute TotalTime from last non-nil split
	total := time.Duration(0)
	for i := s.currentSegmentIndex - 1; i >= 0; i-- {
		segmentID := s.currentRun.LeafSegments[i].ID
		if split, ok := s.currentRun.Splits[segmentID]; ok {
			total = split.CurrentCumulative
			break
		}
	}
	s.currentRun.TotalTime = total

	if s.sessionState == Finished {
		s.sessionState = Running
		s.timer.Start()
	}
}

// Skip sets the current segment to the next one without recording a split
func (s *Service) Skip() {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.sendUpdate()

	if s.currentRun == nil ||
		s.currentSegmentIndex >= len(s.leafSegments)-1 ||
		s.sessionState == Idle ||
		s.sessionState == Finished {
		return
	}
	s.currentSegmentIndex++
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

func (s *Service) SplitFile() (SplitFile, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sf := SplitFile{}
	if s.loadedSplitFile == nil {
		return sf, false
	}
	return deepCopySplitFile(s.loadedSplitFile), true
}

// Dirty returns the unsaved changes status of the session
func (s *Service) Dirty() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.dirty }

// ClearDirty clears th dirty flag, indicating that this session is up to date with what is on repo
func (s *Service) ClearDirty() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dirty = false
}

// State returns the session State
func (s *Service) State() State { s.mu.Lock(); defer s.mu.Unlock(); return s.sessionState }

// Index returns the current segment index of the session
func (s *Service) Index() int { s.mu.Lock(); defer s.mu.Unlock(); return s.currentSegmentIndex }

// Run returns the currently loaded Run
func (s *Service) Run() (Run, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := Run{}
	if s.currentRun == nil {
		return r, false
	}
	r = deepCopyRun(*s.currentRun)
	return r, true
}

// resetLocked assumes that the system is under lock when called.
func (s *Service) resetLocked() {
	s.timer.Pause()
	s.timer.Reset()

	s.currentRun = nil
	s.sessionState = Idle
	s.currentSegmentIndex = -1
}

func (s *Service) PersistRunToSession() {
	if s.currentRun != nil {
		s.loadedSplitFile.Runs = append(s.loadedSplitFile.Runs, *s.currentRun)
		s.loadedSplitFile.BuildStats()
	}
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

	if len(s.leafSegments) == 0 {
		logger.Debug("Split() called on run with no DeepCopyLeafSegments, NO-OP")
		return SplitNoop
	}

	s.timer.SubtractTime(s.loadedSplitFile.Offset)
	s.timer.Start()
	s.loadedSplitFile.Attempts++
	s.sessionState = Running
	s.currentSegmentIndex = 0
	s.currentRun = &Run{
		ID:               uuid.New(),
		Splits:           map[uuid.UUID]Split{},
		LeafSegments:     s.loadedSplitFile.DeepCopyLeafSegments(),
		SplitFileVersion: s.loadedSplitFile.Version,
	}

	s.dirty = true
	return SplitStarted
}

func (s *Service) advanceRun() SplitResult {
	if s.currentSegmentIndex < 0 || s.currentSegmentIndex >= len(s.leafSegments) {
		logger.Warn(
			fmt.Sprintf("Split() called in Running state, but current segment index is out of bounds: %d",
				s.currentSegmentIndex))
		return SplitNoop
	}
	now := s.timer.GetCurrentTime()

	//if splitfile has a negative offset, don't let user split until it starts counting
	if now < 1*time.Millisecond {
		return SplitNoop
	}

	// find prev cumulative from the last non-nil split
	prev := time.Duration(0)
	for i := s.currentSegmentIndex - 1; i >= 0; i-- {
		segmentID := s.currentRun.LeafSegments[i].ID
		if split, ok := s.currentRun.Splits[segmentID]; ok {
			prev = split.CurrentCumulative
			break
		}
	}

	segTime := now - prev
	segmentID := s.currentRun.LeafSegments[s.currentSegmentIndex].ID
	s.currentRun.Splits[segmentID] = Split{
		SplitSegmentID:    segmentID,
		CurrentCumulative: now,
		CurrentDuration:   segTime,
	}

	s.dirty = true
	s.currentSegmentIndex++

	if s.currentSegmentIndex > len(s.leafSegments)-1 {
		s.timer.Pause()
		s.sessionState = Finished
		s.currentRun.TotalTime = now
		s.currentRun.Completed = true
		s.PersistRunToSession()
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

func deepCopySplitFile(inFile *SplitFile) SplitFile {
	segments := deepCopySegments(inFile.Segments)
	runs := deepCopyRuns(inFile.Runs)

	var pbRun *Run
	if inFile.PB != nil {
		pbCopy := deepCopyRun(*inFile.PB)
		pbRun = &pbCopy
	}

	return SplitFile{
		ID:           inFile.ID,
		GameName:     inFile.GameName,
		GameCategory: inFile.GameCategory,
		Version:      inFile.Version,
		Attempts:     inFile.Attempts,
		Offset:       inFile.Offset,
		Segments:     segments,
		WindowX:      inFile.WindowX,
		WindowY:      inFile.WindowY,
		WindowHeight: inFile.WindowHeight,
		WindowWidth:  inFile.WindowWidth,
		SOB:          inFile.SOB,
		Runs:         runs,
		PB:           pbRun,
	}
}

func deepCopyRuns(inRuns []Run) []Run {
	runs := make([]Run, len(inRuns))
	for i := range inRuns {
		runs[i] = deepCopyRun(inRuns[i])
	}
	return runs
}

func deepCopyRun(run Run) Run {
	segments := deepCopySegments(run.LeafSegments)
	splits := deepCopySplits(run.Splits)

	return Run{
		ID:               run.ID,
		SplitFileVersion: run.SplitFileVersion,
		TotalTime:        run.TotalTime,
		Splits:           splits,
		Completed:        run.Completed,
		LeafSegments:     segments,
	}
}

func deepCopySplits(inSplits map[uuid.UUID]Split) map[uuid.UUID]Split {
	splits := map[uuid.UUID]Split{}
	for segmentID, split := range inSplits {
		splits[segmentID] = Split{
			SplitSegmentID:    split.SplitSegmentID,
			CurrentCumulative: split.CurrentCumulative,
			CurrentDuration:   split.CurrentDuration,
		}
	}
	return splits
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
