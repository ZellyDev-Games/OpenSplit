package session

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type MockPersister struct {
	ctx        context.Context
	SaveCalled int
	LoadCalled int
}

func (m *MockPersister) Startup(ctx context.Context, payload *Service) error {
	m.ctx = ctx
	return nil
}

func (m *MockPersister) Load() (SplitFilePayload, error) {
	m.LoadCalled++
	return SplitFilePayload{
		ID:           uuid.New().String(),
		GameName:     "Test Loaded Game",
		GameCategory: "Test Loaded Category",
		Segments: []SegmentPayload{{
			ID:   "037ba872-2fdd-4531-aaee-101d777408b4",
			Name: "Test Loaded Segment",
		}},
		Attempts: 50,
		Runs: []RunPayload{{
			ID:               "037ba872-2fdd-4531-aaee-101d777408b4",
			SplitFileVersion: 1,
		}},
	}, nil
}

func (m *MockPersister) Save(splitFilePayload SplitFilePayload) error {
	m.SaveCalled++
	return nil
}

func getTestFile() *SplitFile {
	return &SplitFile{
		gameName:     "Test Game",
		gameCategory: "Test Category",
		segments: []Segment{{
			id:   uuid.UUID{},
			name: "Test Segment 1",
		}},
		attempts: 0,
	}
}

func TestAttempts(t *testing.T) {
	s := getTestFile()
	s.NewAttempt()
	if s.attempts != 1 {
		t.Errorf("Test failed. Expected 1 attempt, got %d", s.attempts)
	}

	s.SetAttempts(50)
	if s.attempts != 50 {
		t.Errorf("Test failed. Expected 50 attempts, got %d", s.attempts)
	}
}

func TestNewSplitFile(t *testing.T) {
	s, _, _ := getService()
	p := &MockPersister{}
	payload := SplitFilePayload{
		ID:           uuid.New().String(),
		Version:      0,
		GameName:     "Test New Game",
		GameCategory: "Test New Category",
		Segments:     nil,
		Attempts:     0,
		Runs:         nil,
	}

	sf, _ := UpdateSplitFile(nil, p, nil, payload)
	s.SetLoadedSplitFile(sf)
	if s.loadedSplitFile.id == uuid.Nil {
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
	s, _, sf := getService()
	p := &MockPersister{}
	payload := sf.GetPayload()
	payload.GameName = "Updated Game"
	payload.Segments[0].Name = "UPDATED SEGMENT 1"
	existingPayload := s.loadedSplitFile.GetPayload()
	newSF, _ := UpdateSplitFile(nil, p, &existingPayload, payload)
	s.SetLoadedSplitFile(newSF)

	if s.loadedSplitFile.id != sf.id ||
		s.loadedSplitFile.version != sf.version+1 ||
		s.loadedSplitFile.gameName != "Updated Game" ||
		s.loadedSplitFile.gameCategory != sf.gameCategory ||
		s.loadedSplitFile.segments[0].id != sf.segments[0].id ||
		s.loadedSplitFile.segments[1].id != sf.segments[1].id ||
		s.loadedSplitFile.segments[0].name != "UPDATED SEGMENT 1" ||
		s.loadedSplitFile.segments[1].name != sf.segments[1].name {
		t.Fatalf("UpdateSplitFile want %v\ngot\n%v", s.loadedSplitFile, sf)
	}

	// Test unchanged
	newPayload := s.loadedSplitFile.GetPayload()
	_, _ = UpdateSplitFile(nil, p, &newPayload, newPayload)
	if newPayload.Version != s.loadedSplitFile.version {
		t.Error("session UpdateSplitFile bumped version on unchanged file")
	}

	// Test changed
	oldPayload := s.loadedSplitFile.GetPayload()
	newPayload.GameName = "new game"
	sf, _ = UpdateSplitFile(nil, p, &oldPayload, newPayload)
	if sf.gameName != "new game" {
		t.Error("session UpdateSplitFile did not update game name")
	}

	if s.loadedSplitFile.version != 2 {
		t.Error("session UpdateSplitFile bumped splitfile on name only change")
	}

	// Test category change (should bump version)
	newPayload.GameCategory = "a brand new category"
	sf, _ = UpdateSplitFile(nil, p, &oldPayload, newPayload)

	if sf.gameCategory != "a brand new category" {
		t.Error("session UpdateSplitFile did not update game category")
	}

	if sf.version != 3 {
		t.Error("session UpdateSplitFile didn't bump version on category change")
	}
}

func TestLoadSplitFile(t *testing.T) {
	s, _, _ := getService()
	p := &MockPersister{}
	newSF, _ := LoadSplitFile(p)
	s.SetLoadedSplitFile(newSF)

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
