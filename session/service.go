package session

import (
	"OpenSplit/logger"
	"OpenSplit/timer"
	"OpenSplit/utils"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Config struct {
	SpeedRunAPIBase string `json:"speed_run_API_base"`
}

func (s *Service) GetConfig() *Config {
	speedRunBase := os.Getenv("SPEEDRUN_API_BASE")
	if speedRunBase == "" {
		speedRunBase = "https://www.speedrun.com/api/v1"
	}
	return &Config{
		SpeedRunAPIBase: speedRunBase,
	}
}

type ServicePayload struct {
	SplitFile            *SplitFilePayload `json:"split_file"`
	CurrentSegmentIndex  int               `json:"current_segment_index"`
	CurrentSegment       *SegmentPayload   `json:"current_segment"`
	Finished             bool              `json:"finished"`
	Paused               bool              `json:"paused"`
	CurrentTime          time.Duration     `json:"current_time"`
	CurrentTimeFormatted string            `json:"current_time_formatted"`
}

type SplitPayload struct {
	SplitIndex   int            `json:"split_index"`
	NewIndex     int            `json:"new_index"`
	SplitSegment SegmentPayload `json:"split_segment"`
	NewSegment   SegmentPayload `json:"new_segment"`
	Finished     bool           `json:"finished"`
	CurrentTime  string         `json:"current_time"`
}

type Service struct {
	ctx                 context.Context
	timer               *timer.Service
	loadedSplitFile     *SplitFile
	currentSegment      *Segment
	currentSegmentIndex int
	finished            bool
	timeUpdatedChannel  chan time.Duration
	persister           JsonFile
}

func NewService(timer *timer.Service, timeUpdatedChannel chan time.Duration, splitFile *SplitFile, persister JsonFile) *Service {
	service := &Service{
		timer:              timer,
		timeUpdatedChannel: timeUpdatedChannel,
		loadedSplitFile:    splitFile,
		persister:          persister,
	}

	return service
}

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
				runtime.EventsEmit(s.ctx, "timer:update", updatedTime.Milliseconds())
			}
		}
	}()
}

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
		runtime.EventsEmit(s.ctx, "session:update", s.getServicePayload())
		runtime.EventsEmit(s.ctx, "split:update", s.getSplitPayload())
		logger.Debug("split called with last segment in loaded split file, run complete")
		return
	}

	s.currentSegment = &s.loadedSplitFile.segments[s.currentSegmentIndex]
	if s.currentSegmentIndex == 0 {
		s.timer.Reset()
		s.timer.Start()
		s.loadedSplitFile.NewAttempt()
		runtime.EventsEmit(s.ctx, "session:update", s.getServicePayload())
		runtime.EventsEmit(s.ctx, "split:update", s.getSplitPayload())
		logger.Debug(fmt.Sprintf("starting new run (%s - %s - %s) attempt #%d",
			s.loadedSplitFile.gameName,
			s.loadedSplitFile.gameCategory,
			s.currentSegment.name,
			s.loadedSplitFile.attempts))
	} else {
		runtime.EventsEmit(s.ctx, "session:update", s.getServicePayload())
		runtime.EventsEmit(s.ctx, "split:update", s.getSplitPayload())
		logger.Debug(fmt.Sprintf("segment index %d (%s) completed at %s, loading segment %d (%s)",
			s.currentSegmentIndex-1,
			s.loadedSplitFile.segments[s.currentSegmentIndex-1].name,
			s.timer.GetCurrentTimeFormatted(),
			s.currentSegmentIndex,
			s.currentSegment.name))
	}
}

func (s *Service) Pause() {
	if s.timer.IsRunning() {
		s.timer.Pause()
		runtime.EventsEmit(s.ctx, "session:update", s.getServicePayload())
		logger.Debug(fmt.Sprintf("pausing timer at %s", s.timer.GetCurrentTimeFormatted()))
	} else {
		s.timer.Start()
		runtime.EventsEmit(s.ctx, "session:update", s.getServicePayload())
		logger.Debug(fmt.Sprintf("restarting timer at %s", s.timer.GetCurrentTimeFormatted()))
	}
}

func (s *Service) Reset() {
	s.timer.Pause()
	s.timer.Reset()
	s.finished = false
	s.currentSegmentIndex = -1
	s.currentSegment = nil
	runtime.EventsEmit(s.ctx, "timer:update", 0)
	runtime.EventsEmit(s.ctx, "session:update", s.getServicePayload())
	if s.loadedSplitFile != nil {
		logger.Debug(fmt.Sprintf("session reset (%s - %s)", s.loadedSplitFile.gameName, s.loadedSplitFile.gameCategory))
	} else {
		logger.Debug("session reset (no loaded split file)")
	}
}

func (s *Service) UpdateSplitFile(payload SplitFilePayload) {
	newSplitFile, err := newFromPayload(payload)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to parse split file payload: %s", err))
		return
	}

	s.loadedSplitFile = newSplitFile
	err = s.persister.Save(s.loadedSplitFile.GetPayload())
	if err != nil {
		logger.Error(fmt.Sprintf("failed to save split file: %s", err))
		s.loadedSplitFile = nil
	}
}

func (s *Service) LoadSplitFile() (SplitFilePayload, error) {
	newSplitFile, err := s.persister.Load()
	if err != nil {
		s.loadedSplitFile = nil
		return SplitFilePayload{}, err
	}
	return newSplitFile, nil
}

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
