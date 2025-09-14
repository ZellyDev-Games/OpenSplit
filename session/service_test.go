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
	ctx        context.Context
	SaveCalled int
	LoadCalled int
}

func (m *MockPersister) Startup(ctx context.Context) {
	m.ctx = ctx
}

func (m *MockPersister) Load() (SplitFilePayload, error) {
	m.LoadCalled++
	return SplitFilePayload{
		GameName:     "Test Loaded Game",
		GameCategory: "Test Loaded Category",
		Segments: []SegmentPayload{{
			ID:       "037ba872-2fdd-4531-aaee-101d777408b4",
			Name:     "Test Loaded Segment",
			BestTime: "00:00:01.00",
			Average:  "00:00:02.00",
		}},
		Attempts: 50,
	}, nil
}

func (m *MockPersister) Save(splitFilePayload SplitFilePayload, splitFile SplitFile) error {
	m.SaveCalled++
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
			id:          uuid.MustParse("037ba872-2fdd-4531-aaee-101d777408b4"),
			name:        "Test Segment 1",
			bestTime:    time.Second * 1,
			averageTime: time.Second * 2,
		}, {
			id:          uuid.MustParse("4bc1a05c-d4f3-4095-887f-519e2fbb54f3"),
			name:        "Test Segment 2",
			bestTime:    time.Second * 3,
			averageTime: time.Second * 4,
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

func TestPause(t *testing.T) {
	s, mt, _, _ := getService()
	mt.Running = true
	s.Pause()
	if mt.PauseCalled != 1 {
		t.Error("session Pause did not pause timer")
	}

	mt.Running = false
	s.Pause()
	if mt.StartCalled != 1 {
		t.Error("session Pause toggle did not start timer")
	}
}

func TestReset(t *testing.T) {
	s, mt, _, _ := getService()
	s.finished = true
	s.Reset()
	if mt.PauseCalled != 1 {
		t.Error("session Reset did not pause timer")
	}

	if mt.ResetCalled != 1 {
		t.Error("session Reset did not reset timer")
	}

	if s.finished != false {
		t.Error("session Reset did not unflag finished")
	}

	if s.currentSegmentIndex != -1 {
		t.Error("session Reset did not reset segment index")
	}

	if s.currentSegment != nil {
		t.Error("session Reset did not reset current segment")
	}
}

func TestUpdateSplitFile(t *testing.T) {
	s, _, p, sf := getService()
	payload := sf.GetPayload()
	s.loadedSplitFile = nil
	_ = s.UpdateSplitFile(payload)

	if s.loadedSplitFile.gameName != sf.gameName ||
		s.loadedSplitFile.gameCategory != sf.gameCategory ||
		s.loadedSplitFile.segments[0].id != sf.segments[0].id ||
		s.loadedSplitFile.segments[1].id != sf.segments[1].id ||
		s.loadedSplitFile.segments[0].name != sf.segments[0].name ||
		s.loadedSplitFile.segments[1].name != sf.segments[1].name ||
		s.loadedSplitFile.segments[0].bestTime != sf.segments[0].bestTime ||
		s.loadedSplitFile.segments[1].bestTime != sf.segments[1].bestTime ||
		s.loadedSplitFile.segments[0].averageTime != sf.segments[0].averageTime ||
		s.loadedSplitFile.segments[1].averageTime != sf.segments[1].averageTime {
		t.Error("UpdateSplitFile did not set expected splitfile")
	}

	if p.SaveCalled != 1 {
		t.Error("session UpdateSplitFile did not save splitfile")
	}
}

func TestLoadSplitFile(t *testing.T) {
	s, mt, p, _ := getService()
	_, _ = s.LoadSplitFile()

	if p.LoadCalled != 1 {
		t.Error("session LoadSplitFile did not call persister Load")
	}

	if s.loadedSplitFile.gameName != "Test Loaded Game" {
		t.Errorf("load split file game name want: %s, got: %s", "Test Loaded Game", s.loadedSplitFile.gameName)
	}

	if s.loadedSplitFile.gameCategory != "Test Loaded Category" {
		t.Errorf("load split file game category want: %s, got: %s", "Test Loaded Category", s.loadedSplitFile.gameCategory)
	}

	if s.loadedSplitFile.segments[0].id != uuid.MustParse("037ba872-2fdd-4531-aaee-101d777408b4") {
		t.Errorf("load split file segment id want %s got %s", uuid.MustParse("037ba872-2fdd-4531-aaee-101d777408b4"), s.loadedSplitFile.segments[0].id)
	}

	if s.loadedSplitFile.segments[0].name != "Test Loaded Segment" {
		t.Errorf("load split file segment name want: %s, got: %s", "Test Loaded Segment", s.loadedSplitFile.segments[0].name)
	}

	if s.loadedSplitFile.segments[0].bestTime != time.Second*1 {
		t.Errorf("load split file segment bestTime want: %s, got: %s", time.Second*1, s.loadedSplitFile.segments[0].bestTime)
	}

	if s.loadedSplitFile.segments[0].averageTime != time.Second*2 {
		t.Errorf("load split file segment averageTime want: %s, got: %s", time.Second*2, s.loadedSplitFile.segments[0].averageTime)
	}

	if s.loadedSplitFile.attempts != 50 {
		t.Errorf("load split file attempts want: %d, got: %d", 50, s.loadedSplitFile.attempts)
	}

	if mt.PauseCalled != 1 {
		t.Error("load split file did not pause timer")
	}

	if mt.ResetCalled != 1 {
		t.Error("load split file did not reset timer")
	}

	if s.finished != false {
		t.Error("load split file did not unflag finished")
	}

	if s.currentSegmentIndex != -1 {
		t.Error("load split file did not reset segment index")
	}

	if s.currentSegment != nil {
		t.Error("load split file did not reset current segment")
	}
}

func TestGetSessionStatus(t *testing.T) {
	s, _, _, _ := getService()
	payload := s.getServicePayload()
	statusPayload := s.GetSessionStatus()

	if statusPayload.SplitFile.GameName != payload.SplitFile.GameName {
		t.Error("GetSessionStatus did not return expected payload")
	}
}

func TestGetLoadedSplitFile(t *testing.T) {
	s, _, _, _ := getService()
	payload := s.loadedSplitFile.GetPayload()
	loadedPayload := s.GetLoadedSplitFile()

	if loadedPayload == nil || payload.GameName != loadedPayload.GameName {
		t.Error("GetLoadedSplitFile did not return expected payload")
	}
}
