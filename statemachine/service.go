package statemachine

import (
	"context"
	"errors"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/zellydev-games/opensplit/config"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/keyinfo"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/repo"
	"github.com/zellydev-games/opensplit/session"
)

// machine is a private singleton instance of a *Service that represents a state machine.
var machine *Service

// StateID is a compact identifier for a State
type StateID byte

const (
	WELCOME StateID = iota
	NEWFILE
	EDITING
	RUNNING
	CONFIG
)

// RuntimeProvider wraps Wails.runtimeProvider calls to allow for DI for testing.
type RuntimeProvider interface {
	Startup(ctx context.Context)
	SaveFileDialog(runtime.SaveDialogOptions) (string, error)
	OpenFileDialog(runtime.OpenDialogOptions) (string, error)
	MessageDialog(runtime.MessageDialogOptions) (string, error)
	EventsEmit(string, ...any)
	WindowGetSize() (int, int)
	WindowGetPosition() (int, int)
	Quit()
}

type HotkeyProvider interface {
	StartHook(func(data keyinfo.KeyData)) error
	Unhook() error
}

// state implementations can be operated by the Service and do meaningful work, and communicate state to the frontend
// via runtime.EventsEmit
type state interface {
	OnEnter() error
	OnExit() error
	Receive(command dispatcher.Command, payload *string) (dispatcher.DispatchReply, error)
	String() string
	ID() StateID
}

// Service represents a state machine and holds references to all the tools to allow states to do useful work
type Service struct {
	ctx             context.Context
	currentState    state
	sessionService  *session.Service
	repoService     *repo.Service
	runtimeProvider RuntimeProvider
	hotkeyProvider  HotkeyProvider
	configService   *config.Service
}

// InitMachine sets the global singleton, and gives it a friendly default state
func InitMachine(runtimeProvider RuntimeProvider, repoService *repo.Service, sessionService *session.Service, configService *config.Service) *Service {
	machine = &Service{
		sessionService:  sessionService,
		runtimeProvider: runtimeProvider,
		repoService:     repoService,
		configService:   configService,
	}
	return machine
}

// Startup is called by Wails.Run to pass in a context to use against Wails.platform
func (s *Service) Startup(ctx context.Context) {
	machine.ctx = ctx
	machine.changeState(WELCOME, s.sessionService)
}

// AttachHotkeyProvider allows us to receive Dispatch payloads from the given HotkeyProvider
func (s *Service) AttachHotkeyProvider(provider HotkeyProvider) {
	s.hotkeyProvider = provider
}

// ReceiveDispatch allows external facing code to send Command bytes to the state machine
func (s *Service) ReceiveDispatch(command dispatcher.Command, payload *string) (dispatcher.DispatchReply, error) {
	if s.currentState == nil {
		logger.Error("command sent to state machine without a loaded state")
		return dispatcher.DispatchReply{}, errors.New("command sent to state machine without a loaded state")
	}

	if command == dispatcher.QUIT {
		logger.Debug("QUIT command dispatched from front end")
		s.runtimeProvider.Quit()
		return dispatcher.DispatchReply{}, nil
	}

	logger.Debug(fmt.Sprintf("command %d dispatched to state %s", command, s.currentState.String()))
	return s.currentState.Receive(command, payload)
}

// changeState provides a structured way to change the current state, calling appropriate lifecycle methods along the way
func (s *Service) changeState(newState StateID, _ ...interface{}) {
	if s.currentState != nil {
		logger.Debug(fmt.Sprintf("exiting state %s", s.currentState.String()))
		if err := s.currentState.OnExit(); err != nil {
			logger.Error(fmt.Sprintf("OnExit failed: %v", err))
		}
	}

	switch newState {
	case WELCOME:
		logger.Debug("entering state Welcome")
		s.currentState, _ = NewWelcomeState()
	case NEWFILE:
		logger.Debug("entering state NewFile")
		s.currentState, _ = NewNewFileState()
	case EDITING:
		logger.Debug("entering state Editing")
		s.currentState, _ = NewEditingState()
	case RUNNING:
		logger.Debug("entering state Running")
		s.currentState, _ = NewRunningState()
	case CONFIG:
		logger.Debug("entering state Config")
		configState, _ := NewConfigState(s.currentState.ID())
		s.currentState = configState
	default:
		panic("unhandled default case")
	}

	if s.currentState != nil {
		err := s.currentState.OnEnter()
		if err != nil {
			logger.Error(fmt.Sprintf("OnEnter failed: %v", err))
		}
	}
}
