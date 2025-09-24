package session

import (
	"context"
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Repository defines a contract for a repo provider to operate against
type Repository interface {
	Load() (*SplitFile, error)
	Save(*SplitFile) error
	SaveAs(run *SplitFile) error
}

// RuntimeProvider wraps Wails.runtimeProvider calls to allow for DI for testing.
type RuntimeProvider interface {
	Startup(ctx context.Context)
	SaveFileDialog(runtime.SaveDialogOptions) (string, error)
	OpenFileDialog(runtime.OpenDialogOptions) (string, error)
	MessageDialog(runtime.MessageDialogOptions) (string, error)
	EventsEmit(string, ...any)
	Quit()
}

// Timer is an interface that a stopwatch service must implement to be used by session.Service
type Timer interface {
	Startup(context.Context)
	IsRunning() bool
	Run()
	Start()
	Pause()
	Reset()
	GetCurrentTimeFormatted() string
	GetCurrentTime() time.Duration
}

// FileProvider wraps os hooks and file operations to allow DI for testing.
type FileProvider interface {
	WriteFile(string, []byte, os.FileMode) error
	ReadFile(string) ([]byte, error)
	MkdirAll(string, os.FileMode) error
	UserHomeDir() (string, error)
}
