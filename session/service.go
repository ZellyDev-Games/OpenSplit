package session

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zellydev-games/opensplit/config"
	"github.com/zellydev-games/opensplit/logger"
)

const logModule = "session"
const splitDebounce = 120 * time.Millisecond

// SplitResult describes the outcome of a Split operation.
type SplitResult int

const (
	SplitNoop SplitResult = iota
	SplitStarted
	SplitAdvanced
	SplitFinished
	SplitReset
)

// State represents the lifecycle state of a timing session.
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
	Reset(offset *time.Duration)
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
	Icon     string
	Children []Segment
}

// Run is a snapshot of a SplitFile along with additional data to track a run
type Run struct {
	ID               uuid.UUID
	TotalTime        time.Duration
	Splits           map[uuid.UUID]Split
	LeafSegments     []Segment
	Completed        bool
	ForceFinished    bool
	SplitFileVersion int
}

// Service represents the current state of a run.
//
// It is the primary glue that brings together a Timer, a SplitFile, a Run
// history, tracks the status of the current Run / SplitFile, and communicates
// timer updates to the frontend.
type Service struct {
	mu                    sync.Mutex
	timer                 Timer
	loadedSplitFile       *SplitFile
	leafSegments          []*Segment
	currentRun            *Run
	currentSegmentIndex   int
	sessionState          State
	lastSplitTime         time.Time
	dirty                 bool
	runtimeOffsetOverride *time.Duration
	configService         *config.Service
	sessionUpdateChannel  chan *Service
}

// NewService creates a new Service from the passed in components.
func NewService(
	timer Timer,
	cfg *config.Service,
) (*Service, chan *Service) {
	service := &Service{
		timer:                timer,
		currentSegmentIndex:  -1,
		sessionUpdateChannel: make(chan *Service, 128),
		configService:        cfg,
	}

	return service, service.sessionUpdateChannel
}

// UpdateWindowDimensions updates the window dimensions belonging to the
// currently active splitter layout.
//
// Window geometry is persisted independently for vertical and horizontal
// layouts so changing one layout does not overwrite the other.
func (s *Service) UpdateWindowDimensions(x, y, w, h int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.loadedSplitFile == nil {
		return
	}

	window := s.loadedSplitFile.WindowForLayout(
		s.loadedSplitFile.Layout,
	)

	window.X = x
	window.Y = y
	window.Width = w
	window.Height = h

	logger.Debugf(
		logModule,
		"session received %s window dimensions: x:%d y:%d w:%d h:%d",
		s.loadedSplitFile.Layout,
		x,
		y,
		w,
		h,
	)
}

func (s *Service) refreshLeafSegments() {
	if s.loadedSplitFile == nil {
		return
	}

	s.leafSegments = getLeafSegments(
		s.loadedSplitFile.Segments,
		nil,
	)
}

func (sf *SplitFile) InitializeStatistics(window int) {
	if window <= 0 {
		window = 20
	}

	sf.RollingAverageRuns = window
	sf.RebuildStatistics()
}

func (s *Service) SetLoadedSplitFile(sf SplitFile) {
	logger.Debugf(
		logModule,
		"setting loaded splitfile to %s",
		sf.GameName,
	)

	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.sendUpdate()

	copy := DeepCopySplitFile(&sf)
	s.loadedSplitFile = &copy
	s.refreshLeafSegments()

	logger.Infof(
		logModule,
		"%s loaded in session (segments=%d leaf=%d)",
		sf.GameName,
		len(s.loadedSplitFile.Segments),
		len(s.leafSegments),
	)

	window := 20
	if s.configService != nil &&
		s.configService.RollingAverageRuns > 0 {
		window = s.configService.RollingAverageRuns
	}

	s.loadedSplitFile.InitializeStatistics(window)

	s.resetLocked()
	s.currentRun = nil
	s.currentSegmentIndex = -1
	s.sessionState = Idle
	s.dirty = false

	logger.Infof(
		logModule,
		"%s loaded in session (segments total/leaf %d/%d)",
		sf.GameName,
		len(sf.Segments),
		len(s.leafSegments),
	)
}

func (s *Service) ToggleWorldRecordDisplay() (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.loadedSplitFile == nil {
		return false, errors.New("no split file loaded")
	}

	s.loadedSplitFile.WR.Show = !s.loadedSplitFile.WR.Show

	logger.Debugf(
		logModule,
		"world record display=%v",
		s.loadedSplitFile.WR.Show,
	)

	return s.loadedSplitFile.WR.Show, nil
}

// SetLayout sets the active splitter layout.
//
// The layout is part of the loaded split file and therefore survives
// closing and reopening the split file. Each layout has independent
// window dimensions.
func (s *Service) SetLayout(layout string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.loadedSplitFile == nil {
		return errors.New("no split file loaded")
	}

	if layout != "vertical" && layout != "horizontal" {
		return errors.New("invalid splitter layout")
	}

	s.loadedSplitFile.Layout = layout

	window := s.loadedSplitFile.WindowForLayout(layout)

	logger.Debugf(
		logModule,
		"splitter layout=%s window=x:%d y:%d w:%d h:%d",
		layout,
		window.X,
		window.Y,
		window.Width,
		window.Height,
	)

	return nil
}

// SetRuntimeOffsetOverride replaces the configured splitfile offset
// for the current session only.
func (s *Service) SetRuntimeOffsetOverride(offset time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.runtimeOffsetOverride = &offset

	logger.Infof(
		logModule,
		"runtime offset override set to %dms",
		offset.Milliseconds(),
	)

	s.resetLocked()
}

// ClearRuntimeOffsetOverride removes the runtime-only offset override.
func (s *Service) ClearRuntimeOffsetOverride() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.runtimeOffsetOverride = nil

	s.resetLocked()

	logger.Info(
		logModule,
		"runtime offset override cleared",
	)
}

func (s *Service) effectiveOffset() time.Duration {
	if s.runtimeOffsetOverride != nil {
		return *s.runtimeOffsetOverride
	}

	if s.loadedSplitFile == nil {
		return 0
	}

	return s.loadedSplitFile.Offset
}

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

func (s *Service) Undo() {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.sendUpdate()

	if s.currentRun == nil ||
		s.currentSegmentIndex <= 0 ||
		s.sessionState == Idle {
		return
	}

	s.refreshLeafSegments()

	oldSegmentName := "Finished"

	if s.currentSegmentIndex < len(s.leafSegments) {
		oldSegmentName = s.leafSegments[s.currentSegmentIndex].Name
	}

	s.currentSegmentIndex--

	if s.currentSegmentIndex < 0 ||
		s.currentSegmentIndex >= len(s.leafSegments) {
		logger.Errorf(
			logModule,
			"Undo() set currentSegmentIndex outside of bounds %d",
			s.currentSegmentIndex,
		)
		return
	}

	segmentID :=
		s.currentRun.LeafSegments[s.currentSegmentIndex].ID

	segmentName :=
		s.leafSegments[s.currentSegmentIndex].Name

	delete(s.currentRun.Splits, segmentID)

	total := time.Duration(0)

	for i := s.currentSegmentIndex - 1; i >= 0; i-- {
		segmentID := s.currentRun.LeafSegments[i].ID

		if split, ok := s.currentRun.Splits[segmentID]; ok {
			total = split.CurrentCumulative
			break
		}
	}

	s.currentRun.TotalTime = total

	logger.Infof(
		logModule,
		"undo %s: new total time %d - new current segment: %s",
		oldSegmentName,
		total.Milliseconds(),
		segmentName,
	)

	if s.sessionState == Finished {
		s.sessionState = Running
		s.currentRun.Completed = false

		if len(s.loadedSplitFile.Runs) > 0 {
			lastCompletedRun :=
				s.loadedSplitFile.Runs[len(s.loadedSplitFile.Runs)-1]

			if lastCompletedRun.ID == s.currentRun.ID {
				s.loadedSplitFile.Runs =
					s.loadedSplitFile.Runs[:len(s.loadedSplitFile.Runs)-1]
			}
		}

		s.timer.Start()

		logger.Info(
			logModule,
			"finished status cleared",
		)
	}

	s.loadedSplitFile.RebuildStatistics()
}

func (s *Service) Done() SplitResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.sendUpdate()

	if s.sessionState != Running ||
		s.currentRun == nil {
		return SplitNoop
	}

	now := s.timer.GetCurrentTime()

	s.timer.Pause()

	s.sessionState = Finished
	s.currentRun.TotalTime = now
	s.currentRun.Completed = true
	s.currentRun.ForceFinished = true

	s.PersistRunToSession()

	logger.Infof(
		logModule,
		"run force-finished at %d",
		now.Milliseconds(),
	)

	return SplitFinished
}

func (s *Service) UnDone() SplitResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.sendUpdate()

	if s.sessionState != Finished ||
		s.currentRun == nil {
		return SplitNoop
	}

	if !s.currentRun.ForceFinished {
		logger.Warn(
			logModule,
			"Unfinish called on naturally completed run",
		)
		return SplitNoop
	}

	s.currentRun.Completed = false
	s.currentRun.ForceFinished = false

	if len(s.loadedSplitFile.Runs) > 0 {
		last := s.loadedSplitFile.Runs[len(s.loadedSplitFile.Runs)-1]

		if last.ID == s.currentRun.ID {
			s.loadedSplitFile.Runs =
				s.loadedSplitFile.Runs[:len(s.loadedSplitFile.Runs)-1]
		}
	}

	s.loadedSplitFile.RebuildStatistics()

	s.sessionState = Running
	s.timer.Start()

	logger.Info(
		logModule,
		"force-finished run restored",
	)

	return SplitAdvanced
}

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

	oldSegmentName := s.leafSegments[s.currentSegmentIndex].Name

	s.currentSegmentIndex++

	newSegmentName := s.leafSegments[s.currentSegmentIndex].Name

	logger.Infof(
		logModule,
		"skip segment: old segment: %s, new segment: %s",
		oldSegmentName,
		newSegmentName,
	)
}

func (s *Service) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.sendUpdate()

	if s.sessionState != Running &&
		s.sessionState != Paused {
		return
	}

	if s.sessionState == Running {
		s.sessionState = Paused
		s.timer.Pause()

		logger.Infof(
			logModule,
			"session paused at %d",
			s.timer.GetCurrentTime(),
		)
	} else {
		s.sessionState = Running
		s.timer.Start()

		logger.Info(
			logModule,
			"session resumed",
		)
	}
}

func (s *Service) Reset() {
	logger.Info(
		logModule,
		"reset requested",
	)

	s.mu.Lock()
	s.refreshLeafSegments()
	s.resetLocked()
	s.mu.Unlock()

	s.sendUpdate()
}

func (s *Service) CloseRun() {
	s.mu.Lock()

	s.currentRun = nil
	s.runtimeOffsetOverride = nil

	logger.Debug(
		logModule,
		"runtime offset override cleared",
	)

	s.mu.Unlock()

	logger.Info(
		logModule,
		"run closed, resetting session",
	)

	s.resetLocked()
	s.dirty = false
}

func (s *Service) SplitFile() (SplitFile, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sf := SplitFile{}

	if s.loadedSplitFile == nil {
		return sf, false
	}

	return DeepCopySplitFile(s.loadedSplitFile), true
}

func (s *Service) Dirty() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.dirty
}

func (s *Service) ClearDirty() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.dirty = false

	logger.Debug(
		logModule,
		"dirty flag cleared",
	)
}

func (s *Service) State() State {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.sessionState
}

func (s *Service) Index() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.currentSegmentIndex
}

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

func (s *Service) resetLocked() {
	s.timer.Pause()

	logger.Infof(
		logModule,
		"resetLocked runtimeOffset=%v effectiveOffset=%d",
		s.runtimeOffsetOverride,
		s.effectiveOffset().Milliseconds(),
	)

	offset := s.effectiveOffset()

	s.timer.Reset(&offset)

	s.currentRun = nil
	s.sessionState = Idle
	s.currentSegmentIndex = -1

	if s.loadedSplitFile != nil {
		s.loadedSplitFile.RebuildStatistics()
	}

	logger.Info(
		logModule,
		"session reset",
	)
}

func (s *Service) PersistRunToSession() {
	if s.currentRun == nil {
		logger.Warn(
			logModule,
			"persist requested on nil current run",
		)
		return
	}

	window := 20

	if s.configService != nil {
		window = s.configService.RollingAverageRuns
	}

	s.loadedSplitFile.AddRun(
		s.currentRun,
		window,
	)

	logger.Infof(
		logModule,
		"run persisted total=%d attempts=%d",
		s.currentRun.TotalTime.Milliseconds(),
		s.loadedSplitFile.Attempts,
	)
}

func (s *Service) debounced() bool {
	now := time.Now()

	if s.lastSplitTime.Add(splitDebounce).After(now) {
		logger.Warn(
			logModule,
			"split debounced",
		)

		return false
	}

	s.lastSplitTime = now

	return true
}

func (s *Service) startNewRun() SplitResult {
	if s.loadedSplitFile == nil {
		logger.Debug(
			logModule,
			"Split() called with no loaded dto. NO-OP",
		)

		return SplitNoop
	}

	if len(s.leafSegments) == 0 {
		logger.Warn(
			logModule,
			"loaded split file contains no leaf segments",
		)
	}

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

	logger.Infof(
		logModule,
		"new %s %s run started (attempt: %d)",
		s.loadedSplitFile.GameName,
		s.loadedSplitFile.GameCategory,
		s.loadedSplitFile.Attempts,
	)

	return SplitStarted
}

func (s *Service) advanceRun() SplitResult {
	if s.currentSegmentIndex < 0 ||
		s.currentSegmentIndex >= len(s.leafSegments) {
		logger.Warnf(
			logModule,
			"Split() called in Running state, but current segment index is out of bounds: %d",
			s.currentSegmentIndex,
		)

		return SplitNoop
	}

	now := s.timer.GetCurrentTime()

	if now < 1*time.Millisecond {
		return SplitNoop
	}

	prev := time.Duration(0)

	for i := s.currentSegmentIndex - 1; i >= 0; i-- {
		segmentID := s.currentRun.LeafSegments[i].ID

		if split, ok := s.currentRun.Splits[segmentID]; ok {
			prev = split.CurrentCumulative
			break
		}
	}

	segTime := now - prev

	segmentID :=
		s.currentRun.LeafSegments[s.currentSegmentIndex].ID

	segmentName :=
		s.currentRun.LeafSegments[s.currentSegmentIndex].Name

	s.currentRun.Splits[segmentID] = Split{
		SplitSegmentID:    segmentID,
		CurrentCumulative: now,
		CurrentDuration:   segTime,
	}

	s.dirty = true
	s.currentSegmentIndex++

	logger.Infof(
		logModule,
		"split %s at %d",
		segmentName,
		segTime.Milliseconds(),
	)

	if s.currentSegmentIndex > len(s.leafSegments)-1 {
		logger.Info(
			logModule,
			"run complete",
		)

		s.timer.Pause()

		s.sessionState = Finished
		s.currentRun.TotalTime = now
		s.currentRun.Completed = true

		s.PersistRunToSession()

		return SplitFinished
	}

	return SplitAdvanced
}

func (s *Service) sendUpdate() {
	if s.sessionUpdateChannel == nil {
		return
	}

	select {
	case s.sessionUpdateChannel <- s:
	default:
		logger.Debug(
			logModule,
			"session update skipped",
		)
	}
}
