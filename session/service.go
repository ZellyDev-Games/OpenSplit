package session

import (
	"OpenSplit/logger"
	"OpenSplit/timer"
	"context"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ServicePayload struct {
	SplitFile            SplitFilePayload `json:"split_file"`
	CurrentSegmentIndex  int              `json:"current_segment_index"`
	CurrentSegment       *Segment         `json:"current_segment"`
	Finished             bool             `json:"finished"`
	Paused               bool             `json:"paused"`
	CurrentTime          time.Duration    `json:"current_time"`
	CurrentTimeFormatted string           `json:"current_time_formatted"`
}

type SplitPayload struct {
	SplitIndex           int            `json:"split_index"`
	NewIndex             int            `json:"new_index"`
	SplitSegment         SegmentPayload `json:"split_segment"`
	NewSegment           SegmentPayload `json:"new_segment"`
	Finished             bool           `json:"finished"`
	CurrentTime          time.Duration  `json:"current_time"`
	CurrentTimeFormatted string         `json:"current_time_formatted"`
}

type Service struct {
	ctx                 context.Context
	timer               *timer.Service
	loadedSplitFile     *SplitFile
	currentSegment      *Segment
	currentSegmentIndex int
	finished            bool
	timeUpdatedChannel  chan time.Duration
}

func NewService(timer *timer.Service, timeUpdatedChannel chan time.Duration, splitFile *SplitFile) *Service {
	service := &Service{
		timer:              timer,
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

func (s *Service) getServicePayload() ServicePayload {
	payload := ServicePayload{
		CurrentSegmentIndex:  s.currentSegmentIndex,
		CurrentSegment:       s.currentSegment,
		Finished:             s.finished,
		CurrentTime:          s.timer.GetCurrentTime(),
		CurrentTimeFormatted: s.timer.GetCurrentTimeFormatted(),
		Paused:               !s.timer.IsRunning(),
	}

	if s.loadedSplitFile != nil {
		payload.SplitFile = s.loadedSplitFile.GetPayload()
	}

	return payload
}

func (s *Service) getSplitPayload() SplitPayload {
	loadedSplitFileData := s.loadedSplitFile.GetPayload()
	var payload = SplitPayload{
		SplitIndex:           s.currentSegmentIndex - 1,
		NewIndex:             s.currentSegmentIndex,
		Finished:             s.finished,
		CurrentTime:          s.timer.GetCurrentTime(),
		CurrentTimeFormatted: s.timer.GetCurrentTimeFormatted(),
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
