package repo

import (
	"context"
	"os"
	"testing"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type MockFileProvider struct {
	WriteFileCalled int
	MkdirAllCalled  int
}

func (f *MockFileProvider) WriteFile(filename string, data []byte, perm os.FileMode) error {
	f.WriteFileCalled++
	return nil
}

func (f *MockFileProvider) ReadFile(filename string) ([]byte, error) {
	return []byte(`{"game_name":"Final Fight (SNES)","game_category":"Any%","segments":[{"id":"bb846ce5-e710-4ed7-a648-d09d7de8bc73","name":"Streets","best_time":"0:01:01.00","average_time":"0:01:02.03"},{"id":"d6450ae3-6dfe-40ee-bc51-f4ebfd17a960","name":"Subway 2","best_time":"1:01:03.04","average_time":"1:02:03.04"},{"id":"bb866bc5-b452-4556-bd8e-3c74b965573e","name":"Fin","best_time":"2:00:04.05","average_time":"2:01:05.00"}],"attempts":5}`),
		nil
}

func (f *MockFileProvider) MkdirAll(path string, perm os.FileMode) error {
	f.MkdirAllCalled++
	return nil
}

func (f *MockFileProvider) UserHomeDir() (string, error) {
	return "/home/user/zelly", nil
}

type MockRuntimeProvider struct {
	SaveCalled       int
	LoadCalled       int
	EventsEmitCalled int
}

func (m *MockRuntimeProvider) EventsEmit(string, ...any) {
	m.EventsEmitCalled++
}

func (m *MockRuntimeProvider) Quit() {}

func (m *MockRuntimeProvider) SaveFileDialog(runtime.SaveDialogOptions) (string, error) {
	m.SaveCalled++
	return "testfile", nil
}

func (m *MockRuntimeProvider) OpenFileDialog(runtime.OpenDialogOptions) (string, error) {
	m.LoadCalled++
	return "testfile", nil
}

func (m *MockRuntimeProvider) Startup(context.Context) {
}

func (m *MockRuntimeProvider) MessageDialog(options runtime.MessageDialogOptions) (string, error) {
	return "yes", nil
}

func TestSave(t *testing.T) {
	m := &MockRuntimeProvider{}
	f := &MockFileProvider{}
	j := NewJsonFile(m, f)
	err := j.SaveSplitFile([]byte(""))
	if err != nil {
		t.Error(err)
	}

	if m.SaveCalled != 1 {
		t.Error("SaveSplitFile() never opened SaveFileDialog")
	}

	if f.WriteFileCalled != 1 {
		t.Errorf("WriteFile called %d times, expected 1", f.WriteFileCalled)
	}
}

func TestLoad(t *testing.T) {
	m := &MockRuntimeProvider{}
	f := &MockFileProvider{}
	j := NewJsonFile(m, f)
	payload, err := j.LoadSplitFile()
	if err != nil {
		t.Error(err)
	}

	if m.LoadCalled != 1 {
		t.Error("LoadSplitFile() never opened OpenFileDialog")
	}

	want := `{"game_name":"Final Fight (SNES)","game_category":"Any%","segments":[{"id":"bb846ce5-e710-4ed7-a648-d09d7de8bc73","name":"Streets","best_time":"0:01:01.00","average_time":"0:01:02.03"},{"id":"d6450ae3-6dfe-40ee-bc51-f4ebfd17a960","name":"Subway 2","best_time":"1:01:03.04","average_time":"1:02:03.04"},{"id":"bb866bc5-b452-4556-bd8e-3c74b965573e","name":"Fin","best_time":"2:00:04.05","average_time":"2:01:05.00"}],"attempts":5}`

	if string(payload) != want {
		t.Errorf("LoadSplitFile didn't return expected payload")
	}
}
