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

func (r *MockRepository) Load() (*Run, error) {
	return &Run{
		id:               uid,
		splitFileID:      uid,
		splitFileVersion: 1,
		gameName:         "Test Loaded Game",
		gameCategory:     "Loaded Category",
		segments: []Segment{{
			id:      uid,
			name:    "Segment 1",
			gold:    time.Second * 60,
			average: time.Second * 70,
			pb:      time.Second * 50,
		}, {
			id:      uid2,
			name:    "Segment 2",
			gold:    time.Second * 60,
			average: time.Second * 70,
			pb:      time.Second * 50,
		}},
		attempts:  0,
		sob:       StatTime{},
		totalTime: 0,
		splits:    nil,
	}, nil
}

func (m *MockRepository) Save(run *Run) error {
	m.SaveCalled++
	return nil
}

func (m *MockRepository) SaveAs(run *Run) error {
	m.SaveAsCalled++
	return nil
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

func getService() (*Service, *MockTimer, *MockRepository, *SplitFile) {
	t := new(MockTimer)
	m := new(MockRepository)
	sf := getSplitFile()
	return NewService(m, t), t, m, sf
}

func TestSplit(t *testing.T) {
	s, mt, r, _ := getService()
	s.currentRun, _ = r.Load()
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

	if s.currentRun.attempts != 1 {
		t.Fatalf("Split() attempts want %d, got %d", 1, s.currentRun.attempts)
	}

	time.Sleep(splitDebounce + 1*time.Millisecond)
	s.Split()
	if s.currentSegmentIndex != 1 {
		t.Fatalf("Split() s.currentSegmentIndex want %d, got %d", 1, s.currentSegmentIndex)
	}

	if len(s.currentRun.splits) != 1 {
		t.Fatalf("Split() s.currentRun.splits want %d, got %d", 1, len(s.currentRun.splits))
	}

	if s.currentRun.splits[0].splitSegmentID != uid {
		t.Fatalf("Split() 1st recorded split segment ID want %s, got %s", uid.String(), s.currentRun.splits[0].splitSegmentID.String())
	}

	time.Sleep(splitDebounce + 1*time.Millisecond)
	s.Split()
	if s.sessionState != Finished {
		t.Fatalf("Split() s.sessionState want %v, got %v", Finished, s.sessionState)
	}

	if s.timer.IsRunning() {
		t.Fatalf("Split() s.timer.IsRunning() want %v, got %v", false, s.timer.IsRunning())
	}

	totalTime1 := s.currentRun.splits[0].currentDuration
	totalTime2 := s.currentRun.splits[1].currentDuration
	totalTime := totalTime1 + totalTime2
	if s.currentRun.totalTime != totalTime {
		t.Fatalf("Split() final split total time want %d, got %d (%d + %d)", totalTime, s.currentRun.totalTime, totalTime1, totalTime2)
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

	if s.currentRun.attempts != 2 {
		t.Fatalf("Split() attempts want %d, got %d", 1, s.currentRun.attempts)
	}
}

func TestUndo(t *testing.T) {
	s, _, _, _ := getService()
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

	if len(s.currentRun.splits) != 1 {
		t.Fatalf("Undo() on third segment split count want %d, got %d", 1, len(s.currentRun.splits))
	}
}

func TestSkip(t *testing.T) {
	s, _, _, _ := getService()
	s.Split()
	s.Skip()
	if s.currentSegmentIndex != 1 {
		t.Fatalf("Skip() currentSegmentIndex want %d, got %d", 1, s.currentSegmentIndex)
	}

	if len(s.currentRun.splits) != 0 {
		t.Fatalf("Skip() currentRun.splits count want %d, got %d", 0, len(s.currentRun.splits))
	}
}

func TestPause(t *testing.T) {
	s, _, _, _ := getService()
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

func TestLoad(t *testing.T) {
	s, _, _, _ := getService()
	_ = s.Load()
	if s.currentRun.id != uid {
		t.Fatalf("Load() currentRun.id want %s, got %s", uid.String(), s.currentRun.id.String())
	}
}

func TestSave(t *testing.T) {
	s, _, r, _ := getService()
	_ = s.Save()
	s.dirty = true
	if r.SaveCalled != 1 {
		t.Fatalf("Save() repo save called want %d, got %d", 1, r.SaveCalled)
	}

	if s.dirty {
		t.Fatalf("Save() didn't clear dirty flag")
	}
}

func TestSaveAs(t *testing.T) {
	s, _, r, _ := getService()
	_ = s.SaveAs()
	s.dirty = true
	if r.SaveAsCalled != 1 {
		t.Fatalf("SaveAs() repo save called want %d, got %d", 1, r.SaveCalled)
	}

	if s.dirty {
		t.Fatalf("SaveAs() didn't clear dirty flag")
	}
}

func TestReset(t *testing.T) {
	s, mt, _, _ := getService()
	s.Split()
	s.Reset()

	if mt.ResetCalled != 1 {
		t.Fatalf("Reset() timer Reset called want %d, got %d", 1, mt.ResetCalled)
	}

	if s.currentSegmentIndex != -1 {
		t.Fatalf("Reset() current segment index want %d, got %d", -1, s.currentSegmentIndex)
	}

	if s.sessionState != Idle {
		t.Fatalf("Reset() sessionState want %v, got %v", Idle, s.sessionState)
	}
}

func TestDirty(t *testing.T) {
	s, _, _, _ := getService()
	if s.Dirty() {
		t.Fatalf("Dirty() before split want %v, got %v", false, s.Dirty())
	}

	s.Split()

	if !s.Dirty() {
		t.Fatalf("Dirty() after split want %v, got %v", true, s.Dirty())
	}
}

func TestState(t *testing.T) {
	s, _, _, _ := getService()
	if s.State() != Idle {
		t.Fatalf("State() before split want %v, got %v", Idle, s.State())
	}
	s.Split()

	if s.State() != Running {
		t.Fatalf("State() after split want %v, got %v", Running, s.State())
	}
}

func TestIndex(t *testing.T) {
	s, _, _, _ := getService()
	if s.Index() != -1 {
		t.Fatalf("Index() after split want %v, got %v", -1, s.Index())
	}

	s.Split()
	if s.Index() != 0 {
		t.Fatalf("Index() after split want %v, got %v", 0, s.Index())
	}
}

func TestRun(t *testing.T) {
	s, _, _, _ := getService()
	_, clean := s.Run()
	if clean {
		t.Fatalf("Run() returned clean flag with currentRun == nil")
	}
	_ = s.Load()
	r, clean := s.Run()
	if !clean {
		t.Fatalf("Run() returned not clean flag with valid currentRun")
	}

	if r.id != uid {
		t.Fatalf("Run() returned id want %s, got %s", uid.String(), r.id.String())
	}

}
