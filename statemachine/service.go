package statemachine

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/zellydev-games/opensplit/autosplitter"
	"github.com/zellydev-games/opensplit/config"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/keyinfo"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/repo"
	"github.com/zellydev-games/opensplit/repo/adapters"
	"github.com/zellydev-games/opensplit/session"
)

const logModule = "statemachine"

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
	EventsOn(string, func(...any)) func()
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
	ctx                                   context.Context
	splitfileLock                         sync.Mutex
	currentState                          state
	sessionService                        *session.Service
	repoService                           *repo.Service
	runtimeProvider                       RuntimeProvider
	hotkeyProvider                        HotkeyProvider
	configService                         *config.Service
	saveOnWindowDimensionChanges          bool
	unsubscribeFromWindowDimensionChanges func()
	windowHasFocus                        bool
	valueTable                            *autosplitter.ValueTable
	luaEngine                             *autosplitter.Engine
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

func (s *Service) SetValueTable(engine *autosplitter.Engine, table *autosplitter.ValueTable) {
	s.valueTable = table
	s.luaEngine = engine
}

// Startup is called by Wails.Run to pass in a context to use against Wails.platform
func (s *Service) Startup(ctx context.Context) {
	machine.ctx = ctx
	s.unsubscribeFromWindowDimensionChanges = s.setupWindowDimensionListener()
	machine.changeState(WELCOME, s.sessionService)
}

// AttachHotkeyProvider allows us to receive Dispatch payloads from the given HotkeyProvider
func (s *Service) AttachHotkeyProvider(provider HotkeyProvider) {
	s.hotkeyProvider = provider
}

// ReceiveDispatch allows external facing code to send Command bytes to the state machine
func (s *Service) ReceiveDispatch(command dispatcher.Command, payload *string) (dispatcher.DispatchReply, error) {
	if s.currentState == nil {
		logger.Error(logModule, "command sent to state machine without a loaded state")
		return dispatcher.DispatchReply{}, errors.New("command sent to state machine without a loaded state")
	}

	if command == dispatcher.QUIT {
		logger.Debug(logModule, "QUIT command dispatched from front end")
		_ = s.promptDirtySave()
		if s.unsubscribeFromWindowDimensionChanges != nil {
			s.unsubscribeFromWindowDimensionChanges()
		}
		s.runtimeProvider.Quit()
		return dispatcher.DispatchReply{}, nil
	}

	if command == dispatcher.TOGGLEGLOBAL {
		logger.Debug(logModule, "TOGGLEGLOBAL command dispatched from frontend")
		s.configService.GlobalHotkeysActive = !s.configService.GlobalHotkeysActive
		err := machine.repoService.SaveConfig(machine.configService)
		if err != nil {
			message := fmt.Sprintf("error saving config to repo %s", err)
			return dispatcher.DispatchReply{Code: 1, Message: message}, errors.New(message)
		}

		return dispatcher.DispatchReply{
			Message: fmt.Sprintf("%t", s.configService.GlobalHotkeysActive),
		}, nil
	}

	if command == dispatcher.FOCUS {
		if payload == nil {
			return dispatcher.DispatchReply{Code: 1, Message: "focus requires payload of \"true\" or \"false\""}, nil
		}

		s.windowHasFocus = *payload == "true"
		return dispatcher.DispatchReply{}, nil
	}

	logger.Debugf(logModule, "command %d dispatched to state %s", command, s.currentState.String())
	return s.currentState.Receive(command, payload)
}

// changeState provides a structured way to change the current state, calling appropriate lifecycle methods along the way
func (s *Service) changeState(newState StateID, _ ...interface{}) {
	if s.currentState != nil {
		logger.Debugf(logModule, "exiting state %s", s.currentState.String())
		err := s.currentState.OnExit()
		if err != nil {
			logger.Errorf(logModule, "OnExit failed: %v", err)
		}
	}

	switch newState {
	case WELCOME:
		logger.Debug(logModule, "entering state Welcome")
		s.currentState, _ = NewWelcomeState()
	case NEWFILE:
		logger.Debug(logModule, "entering state NewFile")
		s.currentState, _ = NewNewFileState()
	case EDITING:
		logger.Debug(logModule, "entering state Editing")
		s.currentState, _ = NewEditingState()
	case RUNNING:
		logger.Debug(logModule, "entering state Running")
		s.currentState, _ = NewRunningState()
	case CONFIG:
		logger.Debug(logModule, "entering state Config")
		configState, _ := NewConfigState(s.currentState.ID())
		s.currentState = configState
	default:
		panic("unhandled default case")
	}

	if s.currentState != nil {
		err := s.currentState.OnEnter()
		if err != nil {
			logger.Errorf(logModule, "OnEnter failed: %v", err)
		}
	}
}

func (s *Service) saveSplitFile() error {
	s.splitfileLock.Lock()
	defer s.splitfileLock.Unlock()
	sf, loaded := s.sessionService.SplitFile()
	if !loaded {
		msg := "save called without loaded splitfile"
		return errors.New(msg)
	}
	dto := adapters.DomainSplitFileToDTO(sf)
	err := machine.repoService.SaveSplitFile(dto)
	if err != nil {
		return err
	}

	machine.sessionService.ClearDirty()
	return nil
}

func (s *Service) setupWindowDimensionListener() func() {
	return s.runtimeProvider.EventsOn("window:dimensions", func(data ...any) {
		if s.saveOnWindowDimensionChanges {
			logger.Infof(logModule, "Window dimensions have changed: x:%f y:%f w:%f h:%f", data...)

			x := 10
			y := 10
			w := 100
			h := 100

			if f, ok := data[0].(float64); ok {
				x = max(10, int(f))
			}

			if f, ok := data[1].(float64); ok {
				y = max(10, int(f))
			}

			if f, ok := data[2].(float64); ok {
				w = max(100, int(f))
			}

			if f, ok := data[3].(float64); ok {
				h = max(100, int(f))
			}

			err := machine.repoService.SaveSplitFileWindowDimensions(x, y, w, h)
			if err != nil {
				logger.Errorf(logModule, "SaveSplitFileWindowDimensions failed: %v", err)
			}
			machine.sessionService.UpdateWindowDimensions(x, y, w, h)
		}
	})
}

func (s *Service) promptPartialRun() error {
	run, ok := s.sessionService.Run()
	if ok && !run.Completed {
		response, err := s.runtimeProvider.MessageDialog(runtime.MessageDialogOptions{
			Type:          runtime.QuestionDialog,
			Title:         "Add partial run splits to session?",
			Message:       "Do you want to save the splits from this partial run?",
			Buttons:       []string{"Yes", "No"},
			DefaultButton: "Yes",
		})
		if err != nil {
			return err
		}

		if response == "Yes" {
			s.sessionService.PersistRunToSession()
			return nil
		}
	}
	return nil
}

func (s *Service) promptDirtySave() error {
	if s.sessionService.Dirty() {
		response, err := s.runtimeProvider.MessageDialog(runtime.MessageDialogOptions{
			Type:          runtime.QuestionDialog,
			Title:         "Save New Run Data?",
			Message:       "You have unsaved runs, would you like to save them?",
			Buttons:       []string{"Yes", "No"},
			DefaultButton: "Yes",
		})
		if err != nil {
			return err
		}

		if response == "Yes" {
			return s.saveSplitFile()
		}
	}

	return nil
}
