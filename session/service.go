package session

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/utils"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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

// ServicePayload is a snapshot of the session.Service, useful for communicating the state of the service to the frontend
// without exposing internal data.
type ServicePayload struct {
	SplitFile           *SplitFilePayload `json:"split_file"`
	CurrentSegmentIndex int               `json:"current_segment_index"`
	CurrentSegment      *SegmentPayload   `json:"current_segment"`
	Finished            bool              `json:"finished"`
	Paused              bool              `json:"paused"`
	CurrentTime         StatTime          `json:"current_time"`
	CurrentRun          *RunPayload       `json:"current_run"`
}

// Service represents the current state of a run.
//
// It is the primary glue that brings together a Timer, SplitFile, Run history, tracks the status of the
// current Run / SplitFile, and communicates timer updates to the frontend
// If there's one struct that's key to understand in OpenSplit, it's this one.
type Service struct {
	ctx                 context.Context
	timer               Timer
	loadedSplitFile     *SplitFile
	currentSegment      *Segment
	currentSegmentIndex int
	currentRun          *Run
	finished            bool
	timeUpdatedChannel  chan time.Duration
	dirty               bool
	updateCallbacks     []func(context.Context, ServicePayload)
}

// NewService creates a new Service from the passed in components.
//
// Generally in real code splitFile should be nil and will be populated from Service.UpdateSplitFile or Service.LoadSplitFile
// Timer updates will be sent over the timeUpdatedChannel at approximately 60FPS.
func NewService(timer Timer, timeUpdatedChannel chan time.Duration, splitFile *SplitFile) *Service {
	service := &Service{
		timer:               timer,
		timeUpdatedChannel:  timeUpdatedChannel,
		loadedSplitFile:     splitFile,
		currentSegmentIndex: -1,
	}

	return service
}

// AddCallback adds a callback that is invoked when eventsEmit is called with "session:update"
func (s *Service) AddCallback(cb func(context.Context, ServicePayload)) {
	s.updateCallbacks = append(s.updateCallbacks, cb)
}

// Startup is designed to be called by Wails.run OnStartup to supply the proper context.Context that allows the
// session.Service to call Wails runtime functions that do things like open file dialogs.
//
// It also provides the context to the configured Persister so that it may also open file dialogs, calls Reset to ensure
// the state is fresh, and starts a loop to listen for updates from Timer.  These updates are then passed along to the
// frontend to update the visual timer.
func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
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

func (s *Service) CleanQuit(ctx context.Context, persister Persister) bool {
	if s.dirty {
		res, _ := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.QuestionDialog,
			Title:   "Save Split File",
			Message: "Would you like to save your updated runs before exiting?",
		})

		if res == "Yes" {
			logger.Debug("saving split file on exit")
			err := persister.Save(s.loadedSplitFile.GetPayload())
			if err != nil {
				logger.Error(fmt.Sprintf("failed saving split file on exit: %s", err))
			}
		}
	}

	return false
}

// Split advances the state of a Run
//
// Split has several logical branches depending on the state of the Run.  It can start a Run if currentIndex is -1,
// advance to the next split and generate split information if in the middle of a Run, end a Run once the last segment is
// split, and Reset for a new Run if called when the Run is over.
func (s *Service) Split() {
	if s.loadedSplitFile == nil {
		logger.Debug("split called with no split file loaded: NO-OP")
		return
	}

	// TODO: Handle the case where the user just wants a stopwatch (i.e. no segments in a split file, or no split file loaded at all)
	if len(s.loadedSplitFile.segments) == 0 {
		logger.Debug("split called on a split file with no segments: NO-OP")
		return
	}

	if s.finished {
		s.Reset()
		return
	}

	if s.currentSegmentIndex == -1 {
		// run is starting
		s.dirty = true
		s.timer.Reset()
		s.timer.Start()
		s.loadedSplitFile.NewAttempt()
		s.currentRun = &Run{
			id:               uuid.New(),
			splitFileVersion: s.loadedSplitFile.version,
		}

		s.currentSegmentIndex++
		s.currentSegment = &s.loadedSplitFile.segments[s.currentSegmentIndex]
		logger.Debug("sending session update from run start split")
		s.emitEvent("session:update", s.getServicePayload())

		logger.Debug(fmt.Sprintf("starting new run (%s - %s - %s) attempt #%d",
			s.loadedSplitFile.gameName,
			s.loadedSplitFile.gameCategory,
			s.currentSegment.name,
			s.loadedSplitFile.attempts))
		return
	} else if s.currentSegmentIndex >= len(s.loadedSplitFile.segments)-1 {
		// run is finished
		s.timer.Pause()
		s.finished = true
		s.currentRun.splits = append(s.currentRun.splits, Split{
			splitIndex:      s.currentSegmentIndex,
			splitSegmentID:  s.currentSegment.id,
			currentDuration: s.timer.GetCurrentTime(),
		})
		s.currentRun.completed = true
		s.currentRun.totalTime = s.timer.GetCurrentTime()
		logger.Debug("split called with last segment in loaded split file, run complete")
		logger.Debug("sending session update from run complete split")
		s.emitEvent("session:update", s.getServicePayload())
		return
	} else {
		// run is in progress
		s.currentRun.splits = append(s.currentRun.splits, Split{
			splitIndex:      s.currentSegmentIndex,
			splitSegmentID:  s.currentSegment.id,
			currentDuration: s.timer.GetCurrentTime(),
		})
		s.currentSegmentIndex++
		s.currentSegment = &s.loadedSplitFile.segments[s.currentSegmentIndex]
		logger.Debug("sending session update from mid run split")
		s.emitEvent("session:update", s.getServicePayload())
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
		logger.Debug("sending session update from pause timer")
		s.emitEvent("session:update", s.getServicePayload())
		logger.Debug(fmt.Sprintf("pausing timer at %s", s.timer.GetCurrentTimeFormatted()))
	} else {
		s.timer.Start()
		logger.Debug("sending session update from start timer")
		s.emitEvent("session:update", s.getServicePayload())
		logger.Debug(fmt.Sprintf("restarting timer at %s", s.timer.GetCurrentTimeFormatted()))
	}
}

// Reset brings the system back to a default state.
//
// If there was a current Run loaded, information about that Run is added to the SplitFile history.
func (s *Service) Reset() {
	s.timer.Pause()
	s.timer.Reset()

	// If there's a run, add it to the history
	if s.loadedSplitFile != nil && s.currentRun != nil {
		logger.Debug(fmt.Sprintf("appending run to splitfile: %v", s.currentRun))
		s.loadedSplitFile.runs = append(s.loadedSplitFile.runs, *s.currentRun)
		s.loadedSplitFile.BuildStats()
	}

	s.finished = false
	s.currentSegmentIndex = -1
	s.currentSegment = nil
	s.currentRun = nil
	s.emitEvent("timer:update", 0)
	logger.Debug("sending session update from reset session")
	s.emitEvent("session:update", s.getServicePayload())
	if s.loadedSplitFile != nil {
		logger.Debug(fmt.Sprintf("session reset (%s - %s)", s.loadedSplitFile.gameName, s.loadedSplitFile.gameCategory))
	} else {
		logger.Debug("session reset (no loaded split file)")
	}
}

// SetLoadedSplitFile sets the given SplitFile as the loaded one
//
// Splits and other actions only work against the given splitfile
func (s *Service) SetLoadedSplitFile(sf *SplitFile) {
	s.loadedSplitFile = sf
}

// GetSessionStatus is a convenience method for the frontend to query the state of the system imperatively
func (s *Service) GetSessionStatus() ServicePayload {
	return s.getServicePayload()
}

// CloseSplitFile unloads the loaded SplitFile, and resets the system.
func (s *Service) CloseSplitFile() {
	s.loadedSplitFile = nil
	s.Reset()
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
	var loadedSplitFilePayload *SplitFilePayload
	if s.loadedSplitFile != nil {
		payload := s.loadedSplitFile.GetPayload()
		loadedSplitFilePayload = &payload
	}

	var currentSegmentPayload *SegmentPayload
	if s.currentSegment != nil {
		payload := s.currentSegment.GetPayload()
		currentSegmentPayload = &payload
	}

	var currentRunPayload *RunPayload
	if s.currentRun != nil {
		payload := s.currentRun.getPayload()
		currentRunPayload = &payload
	}

	payload := ServicePayload{
		SplitFile:           loadedSplitFilePayload,
		CurrentSegmentIndex: s.currentSegmentIndex,
		CurrentSegment:      currentSegmentPayload,
		Finished:            s.finished,
		CurrentTime: StatTime{
			Raw:       s.timer.GetCurrentTime().Milliseconds(),
			Formatted: utils.FormatTimeToString(s.timer.GetCurrentTime()),
		},
		Paused:     !s.timer.IsRunning(),
		CurrentRun: currentRunPayload,
	}

	return payload
}

// emitEvent wraps the runtime.EventsEmit from Wails so that it no-ops if there is no context.Context provided by
// Wails.run OnStartup callback, a requirement to use the function.  This allows for no-ops in unit testing.
func (s *Service) emitEvent(event string, optional interface{}) {
	if s.ctx != nil {
		if sp, ok := optional.(ServicePayload); ok && event == "session:update" {
			for _, cb := range s.updateCallbacks {
				cb(s.ctx, sp)
			}
		}
		runtime.EventsEmit(s.ctx, event, optional)
	}
}
