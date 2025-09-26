package session

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

var uid = uuid.MustParse("c9bc9698-0f39-488d-80c6-06308f12b03e")
var uid2 = uuid.MustParse("05151851-9132-498e-b70a-344ee03c9384")

type MockTimer struct {
	StartupCalled                 bool
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

func (t *MockTimer) Startup(ctx context.Context) {
	t.StartupCalled = true
}

func (t *MockTimer) Run() {
	t.RunCalled++
}

func (t *MockTimer) Start() {
	t.Running = true
	t.StartCalled++
}

func (t *MockTimer) Pause() {
	t.Running = false
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

type MockRepository struct {
	LoadCalled   int
	SaveCalled   int
	SaveAsCalled int
}

func (r *MockRepository) Load() (*SplitFile, error) {
	return &SplitFile{
		ID:           uid,
		Version:      1,
		GameName:     "Test Loaded Game",
		GameCategory: "Loaded Category",
		Segments: []Segment{{
			ID:      uid,
			Name:    "Segment 1",
			Gold:    time.Second * 60,
			Average: time.Second * 70,
			PB:      time.Second * 50,
		}, {
			ID:      uid2,
			Name:    "Segment 2",
			Gold:    time.Second * 60,
			Average: time.Second * 70,
			PB:      time.Second * 50,
		}},
		Attempts: 0,
		SOB:      0,
	}, nil
}

func (m *MockRepository) Save(*SplitFile) error {
	m.SaveCalled++
	return nil
}

func (m *MockRepository) SaveAs(*SplitFile) error {
	m.SaveAsCalled++
	return nil
}

func getSplitFile() *SplitFile {
	return &SplitFile{
		ID:           uuid.MustParse("037ba872-2fdd-4531-aaee-101d777408b4"),
		GameName:     "Test Game",
		GameCategory: "Test Category",
		Segments: []Segment{{
			ID:   uuid.MustParse("037ba872-2fdd-4531-aaee-101d777408b4"),
			Name: "Test Segment 1",
		}, {
			ID:   uuid.MustParse("4bc1a05c-d4f3-4095-887f-519e2fbb54f3"),
			Name: "Test Segment 2",
		}},
		Attempts: 0,
		Version:  1,
		Runs:     []Run{},
	}
}

func getService() (*Service, *MockTimer, *MockRepository, *SplitFile) {
	t := new(MockTimer)
	m := new(MockRepository)
	sf := getSplitFile()
	service, _ := NewService(t)
	return service, t, m, sf
}

func TestSplit(t *testing.T) {
	s, mt, m, _ := getService()
	s.Split()
	if s.currentSegmentIndex != -1 {
		t.Fatalf("Split() before load s.currentSegmentIndex want %d, got %d", -1, s.currentSegmentIndex)
	}

	sf, _ := m.Load()
	s.SetLoadedSplitFile(sf)
	time.Sleep(splitDebounce + 1*time.Millisecond)
	s.Split()

	if s.currentSegmentIndex != 0 {
		t.Fatalf("Split() s.currentSegmentIndex want %d, got %d", 0, s.currentSegmentIndex)
	}

	if s.dirty == false {
		t.Fatalf("Split() s.dirty want %v, got %v", true, s.dirty)
	}

	if s.sessionState != Running {
		t.Fatalf("Split() s.sessionState want %v, got %v", Running, s.sessionState)
	}

	if !s.timer.IsRunning() {
		t.Fatalf("Split() s.timer.IsRunning() want %v, got %v", true, s.timer.IsRunning())
	}

	if s.loadedSplitFile.Attempts != 1 {
		t.Fatalf("Split() Attempts want %d, got %d", 1, s.loadedSplitFile.Attempts)
	}

	time.Sleep(splitDebounce + 1*time.Millisecond)
	s.Split()
	if s.currentSegmentIndex != 1 {
		t.Fatalf("Split() s.currentSegmentIndex want %d, got %d", 1, s.currentSegmentIndex)
	}

	if len(s.currentRun.Splits) != 1 {
		t.Fatalf("Split() s.currentRun.Splits want %d, got %d", 1, len(s.currentRun.Splits))
	}

	if s.currentRun.Splits[0].SplitSegmentID != uid {
		t.Fatalf("Split() 1st recorded split segment ID want %s, got %s", uid.String(), s.currentRun.Splits[0].SplitSegmentID.String())
	}

	time.Sleep(splitDebounce + 1*time.Millisecond)
	s.Split()
	if s.sessionState != Finished {
		t.Fatalf("Split() s.sessionState want %v, got %v", Finished, s.sessionState)
	}

	if s.timer.IsRunning() {
		t.Fatalf("Split() s.timer.IsRunning() want %v, got %v", false, s.timer.IsRunning())
	}

	totalTime1 := s.currentRun.Splits[0].CurrentDuration
	totalTime2 := s.currentRun.Splits[1].CurrentDuration
	totalTime := totalTime1 + totalTime2
	if s.currentRun.TotalTime != totalTime {
		t.Fatalf("Split() final split total time want %d, got %d (%d + %d)", totalTime, s.currentRun.TotalTime, totalTime1, totalTime2)
	}

	time.Sleep(splitDebounce + 1*time.Millisecond)
	s.Split()
	if s.timer.IsRunning() {
		t.Fatalf("reset Split() timer.IsRunning() want %v, got %v", false, s.timer.IsRunning())
	}

	if mt.ResetCalled != 2 {
		t.Fatalf("Split() timer reset called want %d, got %d", 2, mt.ResetCalled)
	}

	time.Sleep(splitDebounce + 1*time.Millisecond)
	s.Split()
	if s.currentSegmentIndex != 0 {
		t.Fatalf("Split() s.currentSegmentIndex want %d, got %d", 0, s.currentSegmentIndex)
	}

	if s.sessionState != Running {
		t.Fatalf("Split() s.sessionState want %v, got %v", Running, s.sessionState)
	}

	if !s.timer.IsRunning() {
		t.Fatalf("Split() s.timer.IsRunning() want %v, got %v", true, s.timer.IsRunning())
	}

	if s.loadedSplitFile.Attempts != 2 {
		t.Fatalf("Split() Attempts want %d, got %d", 1, s.loadedSplitFile.Attempts)
	}

	if len(s.loadedSplitFile.Runs) != 1 {
		t.Fatalf("Split() Runs count want %d, got %d", 1, len(s.loadedSplitFile.Runs))
	}
}

func TestUndo(t *testing.T) {
	s, _, m, _ := getService()
	sf, _ := m.Load()
	s.SetLoadedSplitFile(sf)
	s.Split()

	// Should do nothing on the first segment
	s.Undo()
	if s.currentSegmentIndex != 0 {
		t.Fatalf("Undo() on first segment currentSegmentIndex want %d, got %d", 0, s.currentSegmentIndex)
	}

	time.Sleep(splitDebounce + 1*time.Millisecond)
	s.Split()

	time.Sleep(splitDebounce + 1*time.Millisecond)
	s.Split()
	s.Undo()
	if s.currentSegmentIndex != 1 {
		t.Fatalf("Undo() on third segment currentSegmentIndex want %d, got %d", 1, s.currentSegmentIndex)
	}

	if len(s.currentRun.Splits) != 1 {
		t.Fatalf("Undo() on third segment split count want %d, got %d", 1, len(s.currentRun.Splits))
	}
}

func TestSkip(t *testing.T) {
	s, _, m, _ := getService()
	sf, _ := m.Load()
	s.SetLoadedSplitFile(sf)
	s.Split()
	s.Skip()
	if s.currentSegmentIndex != 1 {
		t.Fatalf("Skip() currentSegmentIndex want %d, got %d", 1, s.currentSegmentIndex)
	}

	if len(s.currentRun.Splits) != 0 {
		t.Fatalf("Skip() currentRun.Splits count want %d, got %d", 0, len(s.currentRun.Splits))
	}
}

func TestPause(t *testing.T) {
	s, _, m, _ := getService()
	sf, _ := m.Load()
	s.SetLoadedSplitFile(sf)
	s.Pause()
	if s.sessionState != Idle {
		t.Fatalf("Pause() before run start sessionState want %v, got %v", Idle, s.sessionState)
	}

	s.Split()
	s.Pause()
	if s.sessionState != Paused {
		t.Fatalf("Pause() sessionState want %v, got %v", Paused, s.sessionState)
	}

	if s.timer.IsRunning() {
		t.Fatalf("Pause() timer.IsRunning() want %v, got %v", false, s.timer.IsRunning())
	}

	s.Pause()
	if s.sessionState != Running {
		t.Fatalf("Pause() toggle sessionState want %v, got %v", Running, s.sessionState)
	}

	if !s.timer.IsRunning() {
		t.Fatalf("Pause() toggle timer.IsRunning() want %v, got %v", true, s.timer.IsRunning())
	}
}

func TestReset(t *testing.T) {
	s, mt, m, _ := getService()
	sf, _ := m.Load()
	s.SetLoadedSplitFile(sf)
	s.Split()
	s.Reset()

	if mt.ResetCalled != 2 {
		t.Fatalf("Reset() timer Reset called want %d, got %d", 2, mt.ResetCalled)
	}

	if s.currentSegmentIndex != -1 {
		t.Fatalf("Reset() current segment index want %d, got %d", -1, s.currentSegmentIndex)
	}

	if s.sessionState != Idle {
		t.Fatalf("Reset() sessionState want %v, got %v", Idle, s.sessionState)
	}
}

func TestDirty(t *testing.T) {
	s, _, m, _ := getService()
	sf, _ := m.Load()
	s.SetLoadedSplitFile(sf)
	if s.Dirty() {
		t.Fatalf("Dirty() before split want %v, got %v", false, s.Dirty())
	}

	s.Split()

	if !s.Dirty() {
		t.Fatalf("Dirty() after split want %v, got %v", true, s.Dirty())
	}
}

func TestState(t *testing.T) {
	s, _, m, _ := getService()
	sf, _ := m.Load()
	s.SetLoadedSplitFile(sf)
	if s.State() != Idle {
		t.Fatalf("State() before split want %v, got %v", Idle, s.State())
	}
	s.Split()

	if s.State() != Running {
		t.Fatalf("State() after split want %v, got %v", Running, s.State())
	}
}

func TestIndex(t *testing.T) {
	s, _, m, _ := getService()
	sf, _ := m.Load()
	s.SetLoadedSplitFile(sf)
	if s.Index() != -1 {
		t.Fatalf("Index() after split want %v, got %v", -1, s.Index())
	}

	s.Split()
	if s.Index() != 0 {
		t.Fatalf("Index() after split want %v, got %v", 0, s.Index())
	}
}

func TestRun(t *testing.T) {
	s, _, m, _ := getService()
	sf, _ := m.Load()
	s.SetLoadedSplitFile(sf)
	_, clean := s.Run()
	if clean {
		t.Fatalf("Run() returned clean flag with currentRun == nil")
	}
	s.Split()
	r, clean := s.Run()
	if !clean {
		t.Fatalf("Run() returned not clean flag with valid currentRun")
	}

	if r.ID == uuid.Nil {
		t.Fatalf("Run() new run didn't get ID")
	}

}
