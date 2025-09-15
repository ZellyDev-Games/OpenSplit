package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/utils"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Config holds configuration options so that Service.GetConfig can work for both backend and frontend.
type Config struct {
	SpeedRunAPIBase string `json:"speed_run_API_base"`
}

// GetConfig is designed to expose configuration options from the environment or other sources (config files) to the
// frontend.  Go services can just read the environment, but the frontend has no reliable way to do so, so this func
// is bound to the app in main which generates a typescript function for the frontend.
func (s *Service) GetConfig() *Config {
	speedRunBase := os.Getenv("SPEEDRUN_API_BASE")
	if speedRunBase == "" {
		speedRunBase = "https://www.speedrun.com/api/v1"
	}
	return &Config{
		SpeedRunAPIBase: speedRunBase,
	}
}

// ServicePayload is a snapshot of the session.Service, useful for communicating the state of the service to the frontend
// without exposing internal data.
type ServicePayload struct {
	SplitFile            *SplitFilePayload `json:"split_file"`
	CurrentSegmentIndex  int               `json:"current_segment_index"`
	CurrentSegment       *SegmentPayload   `json:"current_segment"`
	Finished             bool              `json:"finished"`
	Paused               bool              `json:"paused"`
	CurrentTime          time.Duration     `json:"current_time"`
	CurrentTimeFormatted string            `json:"current_time_formatted"`
}

// SplitPayload is a snapshot of split data to communicate information about a split to the frontend, and also the
// run history in SplitFile runs
type SplitPayload struct {
	SplitIndex   int            `json:"split_index"`
	NewIndex     int            `json:"new_index"`
	SplitSegment SegmentPayload `json:"split_segment"`
	NewSegment   SegmentPayload `json:"new_segment"`
	Finished     bool           `json:"finished"`
	CurrentTime  string         `json:"current_time"`
}

// Persister is an interface that services that save and load splitfiles must implement to be used by session.Service
type Persister interface {
	Startup(ctx context.Context)
	Load() (split SplitFilePayload, err error)
	Save(split SplitFilePayload) error
}

// Timer is an interface that a stopwatch service must implement to be used by session.Service
type Timer interface {
	IsRunning() bool
	Run()
	Start()
	Pause()
	Reset()
	GetCurrentTimeFormatted() string
	GetCurrentTime() time.Duration
}

// Service represents the interface from the backend Go system to the frontend React system.
//
// It is the primary glue that brings together a Timer, SplitFile, Run history, Persister, and the status of the
// current Run / SplitFile.  If there's one struct that's key to understand in OpenSplit, it's this one.
//
// Service contains the authoritative state of the system, and communicates parts of that state to the front end both by
// imperative functions that are bound to the frontend with Wails.Run, and events sent to the frontend via Service.emitEvent
//
// It communicates timer updates to the frontend, and passes along frontend calls to bound functions to
// the OpenSplit backend systems
type Service struct {
	ctx                 context.Context
	timer               Timer
	loadedSplitFile     *SplitFile
	currentSegment      *Segment
	currentSegmentIndex int
	currentRun          *Run
	finished            bool
	timeUpdatedChannel  chan time.Duration
	persister           Persister
}

// NewService creates a new Service from the passed in components.
//
// Generally in real code splitFile should be nil and will be populated from Service.UpdateSplitFile or Service.LoadSplitFile
// Timer updates will be sent over the timeUpdatedChannel at approximately 60FPS.
func NewService(timer Timer, timeUpdatedChannel chan time.Duration, splitFile *SplitFile, persister Persister) *Service {
	service := &Service{
		timer:               timer,
		timeUpdatedChannel:  timeUpdatedChannel,
		loadedSplitFile:     splitFile,
		persister:           persister,
		currentSegmentIndex: -1,
	}

	return service
}

// Startup is designed to be called by Wails.Run OnStartup to supply the proper context.Context that allows the
// session.Service to call Wails runtime functions that do things like open file dialogs.
//
// It also provides the context to the configured Persister so that it may also open file dialogs, calls Reset to ensure
// the state is fresh, and starts a loop to listen for updates from Timer.  These updates are then passed along to the
// frontend to update the visual timer.
func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
	s.persister.Startup(ctx)
	s.Reset()
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case updatedTime, ok := <-s.timeUpdatedChannel:
				if !ok {
					return
				} // channel closed => exit
				s.emitEvent("timer:update", updatedTime.Milliseconds())
			}
		}
	}()
}

// Split advances the state of a run
//
// Split has several logical branches depending on the state of the run.  It can start a run if currentIndex is -1,
// advance to the next split and generate split information if in the middle of a run, end a run once the last segment is
// split, and Reset for a new run if called when the run is over.
func (s *Service) Split() {
	if s.loadedSplitFile == nil {
		logger.Debug("split called with no split file loaded: NO-OP")
		return
	}

	if s.finished {
		s.Reset()
		return
	} else {
		s.currentSegmentIndex++
	}

	if s.currentSegmentIndex >= len(s.loadedSplitFile.segments) {
		s.timer.Pause()
		s.finished = true
		s.emitEvent("session:update", s.getServicePayload())
		s.emitEvent("session:split", s.getSplitPayload())
		logger.Debug("split called with last segment in loaded split file, run complete")
		return
	}

	s.currentSegment = &s.loadedSplitFile.segments[s.currentSegmentIndex]
	if s.currentSegmentIndex == 0 {
		s.timer.Reset()
		s.timer.Start()
		s.loadedSplitFile.NewAttempt()
		s.currentRun = &Run{
			id:               uuid.New(),
			splitFileID:      s.loadedSplitFile.id,
			splitFileVersion: s.loadedSplitFile.version,
		}
		s.emitEvent("session:update", s.getServicePayload())
		s.emitEvent("session:split", s.getSplitPayload())
		logger.Debug(fmt.Sprintf("starting new run (%s - %s - %s) attempt #%d",
			s.loadedSplitFile.gameName,
			s.loadedSplitFile.gameCategory,
			s.currentSegment.name,
			s.loadedSplitFile.attempts))
	} else {
		s.emitEvent("session:split", s.getSplitPayload())
		logger.Debug(fmt.Sprintf("segment index %d (%s) completed at %s, loading segment %d (%s)",
			s.currentSegmentIndex-1,
			s.loadedSplitFile.segments[s.currentSegmentIndex-1].name,
			s.timer.GetCurrentTimeFormatted(),
			s.currentSegmentIndex,
			s.currentSegment.name))
	}
}

// Pause toggles the timer between the running and not running state.
func (s *Service) Pause() {
	if s.timer.IsRunning() {
		s.timer.Pause()
		s.emitEvent("session:update", s.getServicePayload())
		logger.Debug(fmt.Sprintf("pausing timer at %s", s.timer.GetCurrentTimeFormatted()))
	} else {
		s.timer.Start()
		s.emitEvent("session:update", s.getServicePayload())
		logger.Debug(fmt.Sprintf("restarting timer at %s", s.timer.GetCurrentTimeFormatted()))
	}
}

// Reset brings the system back to a default state.
//
// If there was a current run loaded, information about that run is added to the SplitFile history.
func (s *Service) Reset() {
	s.timer.Pause()
	s.timer.Reset()

	// If there's a run, add it to the history
	if s.loadedSplitFile != nil && s.currentRun != nil {
		s.loadedSplitFile.runs = append(s.loadedSplitFile.runs, *s.currentRun)
	}

	s.finished = false
	s.currentSegmentIndex = -1
	s.currentSegment = nil
	s.emitEvent("timer:update", 0)
	s.emitEvent("session:update", s.getServicePayload())
	if s.loadedSplitFile != nil {
		logger.Debug(fmt.Sprintf("session reset (%s - %s)", s.loadedSplitFile.gameName, s.loadedSplitFile.gameCategory))
	} else {
		logger.Debug("session reset (no loaded split file)")
	}
}

// UpdateSplitFile uses the configured Persister to save the SplitFile to the configured storage.
//
// It creates a SplitFile from the given SplitFilePayload and then sets that SplitFile as the currently loaded one.
func (s *Service) UpdateSplitFile(payload SplitFilePayload) error {
	newSplitFile, err := newFromPayload(payload)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to parse split file payload: %s", err))
		return err
	}

	if s.loadedSplitFile != nil && !SplitFileChanged(payload, s.loadedSplitFile.GetPayload()) {
		logger.Debug("SplitFile unchanged")
		return nil
	}

	s.loadedSplitFile = newSplitFile
	s.loadedSplitFile.version++
	err = s.persister.Save(s.loadedSplitFile.GetPayload())
	if err != nil {
		var cancelled = &UserCancelledSave{}
		if errors.As(err, cancelled) {
			logger.Debug("user cancelled save")
			return err
		}
		logger.Error(fmt.Sprintf("failed to save split file: %s", err))
		s.loadedSplitFile = nil
		return err
	}

	s.emitEvent("splitfile:update", s.loadedSplitFile.GetPayload())
	return err
}

// LoadSplitFile retrieves a SplitFilePayload from Persister configured storage.
//
// It creates a new SplitFile from the retrieved SplitFilePayload, sets that as the loaded split file, and resets the
// system.
func (s *Service) LoadSplitFile() (SplitFilePayload, error) {
	newSplitFilePayload, err := s.persister.Load()
	if err != nil {
		s.loadedSplitFile = nil
		return SplitFilePayload{}, err
	}

	newSplitFile, err := newFromPayload(newSplitFilePayload)
	if err != nil {
		s.loadedSplitFile = nil
		return SplitFilePayload{}, err
	}

	s.loadedSplitFile = newSplitFile
	s.Reset()
	s.emitEvent("splitfile:update", s.loadedSplitFile.GetPayload())
	return newSplitFilePayload, nil
}

// GetSessionStatus is a convenience method for the frontend to query the state of the system imperatively
func (s *Service) GetSessionStatus() ServicePayload {
	return s.getServicePayload()
}

// CloseSplitFile unloads the loaded SplitFile, and resets the system.
func (s *Service) CloseSplitFile() {
	s.loadedSplitFile = nil
	s.Reset()
	s.emitEvent("splitfile:update", nil)
}

// GetLoadedSplitFile returns the SplitFilePayload representation of the currently loaded SplitFile
//
// It returns a payload, modifications to it do not affect the internal state.  To do that modify the payload then
// send the modified payload to UpdateSplitFile.
func (s *Service) GetLoadedSplitFile() *SplitFilePayload {
	if s.loadedSplitFile != nil {
		splitFilePayload := s.loadedSplitFile.GetPayload()
		return &splitFilePayload
	}
	return nil
}

func (s *Service) getServicePayload() ServicePayload {
	var loadedSplitFile *SplitFilePayload
	if s.loadedSplitFile != nil {
		payload := s.loadedSplitFile.GetPayload()
		loadedSplitFile = &payload
	}

	var currentSegmentPayload *SegmentPayload
	if s.currentSegment != nil {
		payload := s.currentSegment.GetPayload()
		currentSegmentPayload = &payload
	}

	payload := ServicePayload{
		SplitFile:            loadedSplitFile,
		CurrentSegmentIndex:  s.currentSegmentIndex,
		CurrentSegment:       currentSegmentPayload,
		Finished:             s.finished,
		CurrentTime:          s.timer.GetCurrentTime(),
		CurrentTimeFormatted: s.timer.GetCurrentTimeFormatted(),
		Paused:               !s.timer.IsRunning(),
	}

	return payload
}

func (s *Service) getSplitPayload() SplitPayload {
	loadedSplitFileData := s.loadedSplitFile.GetPayload()
	var payload = SplitPayload{
		SplitIndex:  s.currentSegmentIndex - 1,
		NewIndex:    s.currentSegmentIndex,
		Finished:    s.finished,
		CurrentTime: utils.FormatTimeToString(s.timer.GetCurrentTime()),
	}

	if !s.finished {
		payload.NewSegment = loadedSplitFileData.Segments[s.currentSegmentIndex]
		payload.NewIndex = s.currentSegmentIndex
	}

	if s.currentSegmentIndex != 0 {
		payload.SplitSegment = loadedSplitFileData.Segments[s.currentSegmentIndex-1]
		payload.SplitIndex = s.currentSegmentIndex - 1
	}

	return payload
}

// emitEvent wraps the runtime.EventsEmit from Wails so that it no-ops if there is no context.Context provided by
// Wails.Run OnStartup callback, a requirement to use the function.  This allows for no-ops in unit testing.
func (s *Service) emitEvent(event string, optional interface{}) {
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, event, optional)
	}
}
