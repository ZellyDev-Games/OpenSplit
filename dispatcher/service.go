package dispatcher

import (
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/zellydev-games/opensplit/logger"
)

const logModule = "dispatcher"

// RuntimeProvider wraps Wails.runtimeProvider calls to allow for DI for testing.
type RuntimeProvider interface {
	OpenFileDialog(runtime.OpenDialogOptions) (string, error)
}

// Command bytes are sent to the Service.Dispatch method receiver to indicate the state machine should take some action.
type Command byte

const (
	QUIT Command = iota
	NEW
	LOAD
	EDIT
	CANCEL
	SUBMIT
	CLOSE
	RESET
	SAVE
	SPLIT
	UNDO
	SKIP
	PAUSE
	TOGGLEGLOBAL
	FOCUS
)

// DispatchReply is sent in response to Dispatch
//
// Code greater than zero indicates an error situation
type DispatchReply struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type DispatchReceiver interface {
	ReceiveDispatch(Command, *string) (DispatchReply, error)
}

type Service struct {
	mu                    sync.Mutex
	receiver              DispatchReceiver
	runtime               RuntimeProvider
	autoSplitterDirectory string
}

func NewService(receiver DispatchReceiver, runtime RuntimeProvider, autoSplitterFileDirectory string) *Service {
	return &Service{
		autoSplitterDirectory: autoSplitterFileDirectory,
		runtime:               runtime,
		receiver:              receiver,
	}
}

func (s *Service) Dispatch(command Command, payload *string) (DispatchReply, error) {
	logger.Debugf(logModule, "dispatching command: %v", command)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.receiver.ReceiveDispatch(command, payload)
}

func (s *Service) PickAutoSplitterFile() (string, error) {
	return s.runtime.OpenFileDialog(runtime.OpenDialogOptions{
		DefaultDirectory: s.autoSplitterDirectory,
		Title:            "Select an Autosplitter lua file",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Autosplitter Files",
				Pattern:     "*.lua",
			},
		},
	})
}
