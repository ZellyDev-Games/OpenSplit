package session

import (
	"OpenSplit/logger"
	"OpenSplit/timer"
	"context"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Persister interface {
	Save(id string, splitFile *SplitFile) error
	Load(id string) (*SplitFile, error)
}

type ServicePayload struct {
	SplitFile            *SplitFile    `json:"split_file"`
	CurrentSegmentIndex  int           `json:"current_segment_index"`
	CurrentSegment       *Segment      `json:"current_segment"`
	Finished             bool          `json:"finished"`
	Paused               bool          `json:"paused"`
	CurrentTime          time.Duration `json:"current_time"`
	CurrentTimeFormatted string        `json:"current_time_formatted"`
}

type SplitPayload struct {
	SplitIndex           int           `json:"split_index"`
	NewIndex             int           `json:"new_index"`
	SplitSegment         *Segment      `json:"split_segment"`
	NewSegment           *Segment      `json:"new_segment"`
	Finished             bool          `json:"finished"`
	CurrentTime          time.Duration `json:"current_time"`
	CurrentTimeFormatted string        `json:"current_time_formatted"`
}

type Service struct {
	ctx                 context.Context
	timer               *timer.Service
	persister           Persister
	loadedSplitFile     *SplitFile
	currentSegment      *Segment
	currentSegmentIndex int
	finished            bool
	timeUpdatedChannel  chan time.Duration
}

func NewService(timer *timer.Service, timeUpdatedChannel chan time.Duration, persister Persister, splitFile *SplitFile) *Service {
	service := &Service{
		timer:              timer,
		persister:          persister,
		timeUpdatedChannel: timeUpdatedChannel,
		loadedSplitFile:    splitFile,
	}

	return service
}

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
				runtime.EventsEmit(s.ctx, "timer:update", updatedTime.Milliseconds())
			}
		}
	}()
}

func (s *Service) Save(id string, splitFile *SplitFile) error {
	return s.persister.Save(id, splitFile)
}

func (s *Service) Load(id string) error {
	splitFile, err := s.persister.Load(id)
	if err != nil {
		return err
	}
	s.loadedSplitFile = splitFile
	s.Reset()
	return nil
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
		runtime.EventsEmit(s.ctx, "session:update", s.getServicePayload())
		runtime.EventsEmit(s.ctx, "split:update", s.getSplitPayload())
		logger.Debug(fmt.Sprintf("starting new run (%s - %s - %s) ",
			s.loadedSplitFile.gameName,
			s.loadedSplitFile.gameCategory,
			s.currentSegment.name))
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
	logger.Debug(fmt.Sprintf("session reset (%s - %s)", s.loadedSplitFile.gameName, s.loadedSplitFile.gameCategory))
}

func (s *Service) getServicePayload() ServicePayload {
	return ServicePayload{
		SplitFile:            s.loadedSplitFile,
		CurrentSegmentIndex:  s.currentSegmentIndex,
		CurrentSegment:       s.currentSegment,
		Finished:             s.finished,
		CurrentTime:          s.timer.GetCurrentTime(),
		CurrentTimeFormatted: s.timer.GetCurrentTimeFormatted(),
		Paused:               !s.timer.IsRunning(),
	}
}

func (s *Service) getSplitPayload() SplitPayload {
	var payload = SplitPayload{
		SplitIndex:           s.currentSegmentIndex - 1,
		NewIndex:             s.currentSegmentIndex,
		Finished:             s.finished,
		CurrentTime:          s.timer.GetCurrentTime(),
		CurrentTimeFormatted: s.timer.GetCurrentTimeFormatted(),
	}

	if !s.finished {
		payload.NewSegment = &s.loadedSplitFile.segments[s.currentSegmentIndex]
		payload.NewIndex = s.currentSegmentIndex
	}

	if s.currentSegmentIndex != 0 {
		payload.SplitSegment = &s.loadedSplitFile.segments[s.currentSegmentIndex-1]
		payload.SplitIndex = s.currentSegmentIndex - 1
	}

	return payload
}
