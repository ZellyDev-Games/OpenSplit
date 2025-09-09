package session

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type MockTimer struct {
	Running                       bool
	IsRunningCalled               int
	RunCalled                     int
	StartCalled                   int
	PauseCalled                   int
	ResetCalled                   int
	GetCurrentTimeFormattedCalled int
	GetCurrentTimeCalled          int
}

func (t *MockTimer) IsRunning() bool {
	t.IsRunningCalled++
	return t.Running
}

func (t *MockTimer) Run() {
	t.RunCalled++
}

func (t *MockTimer) Start() {
	t.StartCalled++
}

func (t *MockTimer) Pause() {
	t.PauseCalled++
}

func (t *MockTimer) Reset() {
	t.ResetCalled++
}

func (t *MockTimer) GetCurrentTimeFormatted() string {
	t.GetCurrentTimeFormattedCalled++
	return "1:02:03.04"
}

func (t *MockTimer) GetCurrentTime() time.Duration {
	t.GetCurrentTimeCalled++
	return time.Hour*1 + time.Minute*2 + time.Second*3 + time.Millisecond*40
}

type MockPersister struct {
	ctx context.Context
}

func (m *MockPersister) Startup(ctx context.Context) {
	m.ctx = ctx
}

func (m *MockPersister) Load() (SplitFilePayload, error) {
	return SplitFilePayload{}, nil
}

func (m *MockPersister) Save(splitFile SplitFilePayload) error {
	return nil
}

func getService() (*Service, *MockTimer, *MockPersister, *SplitFile) {
	t := new(MockTimer)
	p := new(MockPersister)
	sf := getSplitFile()
	return NewService(t, nil, sf, p), t, p, sf
}

func getSplitFile() *SplitFile {
	return &SplitFile{
		gameName:     "Test Game",
		gameCategory: "Test Category",
		segments: []Segment{{
			id:          uuid.UUID{},
			name:        "Test Segment 1",
			bestTime:    1,
			averageTime: 2,
		}, {
			id:          uuid.UUID{},
			name:        "Test Segment 2",
			bestTime:    3,
			averageTime: 4,
		}},
		attempts: 0,
	}
}

func TestServiceSplitWithNoFileLoaded(t *testing.T) {
	s, _, _, _ := getService()
	s.loadedSplitFile = nil
	s.Split()
	if s.currentSegmentIndex != -1 {
		t.Error("Split increased segment index with no splitfile loaded")
	}
}

func TestServiceSplit(t *testing.T) {
	s, mt, _, sf := getService()

	// first split
	s.Split()
	if s.currentSegmentIndex != 0 {
		t.Error("Split did not increment segment index")
	}

	if s.currentSegment != &sf.segments[0] {
		t.Error("Split did not set current segment")
	}

	if mt.ResetCalled != 1 {
		t.Error("first split did not reset timer")
	}

	if mt.StartCalled != 1 {
		t.Error("first split did not start timer")
	}

	if sf.attempts != 1 {
		t.Error("first split did not increment attempts")
	}

	// second split
	s.Split()
	if s.currentSegmentIndex != 1 {
		t.Error("second Split did not increment segment index")
	}

	if s.currentSegment != &sf.segments[1] {
		t.Error("second Split did not set current segment")
	}

	if mt.ResetCalled != 1 {
		t.Error("second Split erroneously reset timer")
	}

	if mt.StartCalled != 1 {
		t.Error("second Split erroneously started timer")
	}

	if sf.attempts != 1 {
		t.Error("second Split erroneously incremented attempts")
	}

	// end split
	s.Split()
	if s.currentSegmentIndex != 2 {
		t.Error("end Split didn't increment segment index")
	}

	if s.currentSegment != &sf.segments[1] {
		t.Error("end Split erroneously changed current segment")
	}

	if mt.PauseCalled != 1 {
		t.Error("end Split did not pause timer")
	}

	if s.finished != true {
		t.Error("end Split did not finish session")
	}

	// reset split
	s.Split()
	if mt.PauseCalled != 2 {
		t.Error("reset Split did not pause timer")
	}

	if mt.ResetCalled != 2 {
		t.Error("reset Split did not reset timer")
	}

	if s.finished != false {
		t.Error("reset Split did not unflag finished")
	}

	if s.currentSegmentIndex != -1 {
		t.Error("reset Split did not reset segment index")
	}

	if s.currentSegment != nil {
		t.Error("reset Split did not reset current segment")
	}

	// first split, new attempt
	s.Split()
	if s.currentSegmentIndex != 0 {
		t.Error("new attempt Split did not increment segment index")
	}

	if s.currentSegment != &sf.segments[0] {
		t.Error("new attempt Split did not set current segment")
	}

	if mt.ResetCalled != 3 {
		t.Error("new attempt split did not reset timer")
	}

	if mt.StartCalled != 2 {
		t.Error("new attempt split did not start timer")
	}

	if sf.attempts != 2 {
		t.Error("new attempt split did not increment attempts")
	}
}
