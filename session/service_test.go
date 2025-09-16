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
			ID:   "037ba872-2fdd-4531-aaee-101d777408b4",
			Name: "Test Loaded Segment",
		}},
		Attempts: 50,
		Runs: []RunPayload{{
			ID:               uuid.MustParse("037ba872-2fdd-4531-aaee-101d777408b4"),
			SplitFileVersion: 1,
		}},
	}, nil
}

func (m *MockPersister) Save(splitFilePayload SplitFilePayload) error {
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
		id:           uuid.MustParse("037ba872-2fdd-4531-aaee-101d777408b4"),
		gameName:     "Test Game",
		gameCategory: "Test Category",
		segments: []Segment{{
			id:   uuid.MustParse("037ba872-2fdd-4531-aaee-101d777408b4"),
			name: "Test Segment 1",
		}, {
			id:   uuid.MustParse("4bc1a05c-d4f3-4095-887f-519e2fbb54f3"),
			name: "Test Segment 2",
		}},
		attempts: 0,
		version:  1,
		runs:     []Run{},
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

	if s.currentRun != nil || len(sf.runs) != 0 {
		t.Error("new Service wasn't started clean")
	}

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

	if s.currentRun == nil {
		t.Error("first split did not set current run")
	}

	if len(sf.runs) > 0 {
		t.Error("first split on new file added run prematurely")
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
	if s.currentSegmentIndex != 1 {
		t.Error("end Split incremented segment index out of range")
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

	if s.currentRun.completed != true {
		t.Error("end Split did not set completed flag on currentRun")
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

	if len(sf.runs) != 1 {
		t.Error("reset did not add finished run to splitfile")
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

	if &sf.runs[0] == s.currentRun {
		t.Error("new attempt split did not set a new run")
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

func TestNewSplitFile(t *testing.T) {
	s, _, _, _ := getService()
	payload := SplitFilePayload{
		ID:           uuid.UUID{},
		Version:      0,
		GameName:     "Test New Game",
		GameCategory: "Test New Category",
		Segments:     nil,
		Attempts:     0,
		Runs:         nil,
	}

	s.loadedSplitFile = nil
	_ = s.UpdateSplitFile(payload)
	emptyUUID := uuid.UUID{}
	if s.loadedSplitFile.id == emptyUUID {
		t.Error("session UpdateSplitFile did not create a new id with a new splitfile")
	}

	if s.loadedSplitFile.version != 1 {
		t.Error("session UpdateSplitFile did not bump version on new file")
	}

	if len(s.loadedSplitFile.runs) > 0 {
		t.Error("session UpdateSplitFile erroneously added runs to a new file")
	}
}

func TestUpdateSplitFile(t *testing.T) {
	s, _, p, sf := getService()
	payload := sf.GetPayload()
	payload.GameName = "Updated Game"
	payload.Segments[0].Name = "UPDATED SEGMENT 1"
	_ = s.UpdateSplitFile(payload)

	if s.loadedSplitFile.id != sf.id ||
		s.loadedSplitFile.version != sf.version+1 ||
		s.loadedSplitFile.gameName != "Updated Game" ||
		s.loadedSplitFile.gameCategory != sf.gameCategory ||
		s.loadedSplitFile.segments[0].id != sf.segments[0].id ||
		s.loadedSplitFile.segments[1].id != sf.segments[1].id ||
		s.loadedSplitFile.segments[0].name != "UPDATED SEGMENT 1" ||
		s.loadedSplitFile.segments[1].name != sf.segments[1].name {
		t.Errorf("UpdateSplitFile want %v\ngot\n%v", s.loadedSplitFile, sf)
	}

	if p.SaveCalled != 1 {
		t.Error("session UpdateSplitFile did not save splitfile")
	}

	if s.loadedSplitFile.version != 2 {
		t.Error("session UpdateSplitFile did not bump new splitfile version on change")
	}

	// Test unchanged
	newPayload := s.loadedSplitFile.GetPayload()
	_ = s.UpdateSplitFile(newPayload)
	if newPayload.Version != s.loadedSplitFile.version {
		t.Error("session UpdateSplitFile bumped version on unchanged file")
	}

	// Test changed
	newPayload.GameName = "new game"
	_ = s.UpdateSplitFile(newPayload)
	if p.SaveCalled != 3 {
		t.Error("session UpdateSplitFile did not save splitfile on change")
	}

	if s.loadedSplitFile.version != 2 {
		t.Error("session UpdateSplitFile did not bump splitfile version on change")
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

	if s.loadedSplitFile.runs[0].id != uuid.MustParse("037ba872-2fdd-4531-aaee-101d777408b4") {
		t.Error("Load split file did not load runs")
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
