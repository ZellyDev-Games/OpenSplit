package statemachine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/session"
)

// machine is a private singleton instance of a *Service that represents a state machine.
var machine *Service

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
)

// StateID is a compact identifier for a State
type StateID byte

const (
	WELCOME StateID = iota
	NEWFILE
	EDITING
	RUNNING
)

// DispatchReply is sent in response to Dispatch
//
// Code greater than zero indicates an error situation
type DispatchReply struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// state implementations can be operated by the Service and do meaningful work, and communicate state to the frontend
// via runtime.EventsEmit
type state interface {
	OnEnter(ctx context.Context) error
	OnExit() error
	Receive(command Command, payload []byte) (DispatchReply, error)
	String() string
}

// Service represents a state machine and holds references to all the tools to allow states to do useful work
type Service struct {
	ctx             context.Context
	currentState    state
	sessionService  *session.Service
	persister       session.Persister
	runtimeProvider session.RuntimeProvider
}

// StartMachine sets the global singleton, and gives it a friendly default state
func StartMachine(runtimeProvider session.RuntimeProvider, sessionService *session.Service, persister session.Persister) *Service {
	machine = &Service{
		sessionService:  sessionService,
		persister:       persister,
		runtimeProvider: runtimeProvider,
	}

	return machine
}

// Startup is called by Wails.Run to pass in a context to use against Wails.runtime
func (s *Service) Startup(ctx context.Context) {
	s.ctx = ctx
	s.sessionService.Startup(ctx)
	s.runtimeProvider.Startup(ctx)
	err := s.persister.Startup(ctx, s.sessionService)
	if err != nil {
		logger.Error("Session Service failed to Startup persister: " + err.Error())
		os.Exit(3)
	}

	machine.changeState(WELCOME, s.sessionService)
}

// Dispatch allows external facing code to send Command bytes to the state machine
func (s *Service) Dispatch(command Command, payload *string) (DispatchReply, error) {
	if s.currentState == nil {
		logger.Error("command sent to state machine without a loaded state")
		return DispatchReply{}, errors.New("command sent to state machine without a loaded state")
	}

	if command == RESET {
		logger.Debug("RESET command dispatched from front end")
		s.changeState(WELCOME)
		return DispatchReply{}, nil
	}

	if command == QUIT {
		logger.Debug("QUIT command dispatched from front end")
		runtime.Quit(s.ctx)
		return DispatchReply{}, nil
	}

	logger.Debug(fmt.Sprintf("%d command dispatched to state %s", command, s.currentState.String()))
	if payload != nil {
		return s.currentState.Receive(command, []byte(*payload))
	} else {
		return s.currentState.Receive(command, nil)
	}
}

// changeState provides a structured way to change the current state, calling appropriate lifecycle methods along the way
func (s *Service) changeState(newState StateID, context ...interface{}) {
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
	default:
		panic("unhandled default case")
	}

	if s.currentState != nil {
		err := s.currentState.OnEnter(s.ctx)
		if err != nil {
			logger.Error(fmt.Sprintf("OnEnter failed: %v", err))
		}
	}
}

// Welcome greets the user by indicating the frontend should display the Welcome screen
type Welcome struct {
	ctx context.Context
}

// NewWelcomeState requires a *session.Service as its one and only payload parameter
func NewWelcomeState() (*Welcome, error) {
	return &Welcome{}, nil
}

// String returns the human friendly name for a State
func (w *Welcome) String() string {
	return "Welcome"
}

// OnEnter sets the context from the Wails app and signals the frontend to show the Welcome component
func (w *Welcome) OnEnter(ctx context.Context) error {
	w.ctx = ctx
	runtime.EventsEmit(ctx, "state:enter", WELCOME)
	return nil
}
func (w *Welcome) OnExit() error { return nil }
func (w *Welcome) Receive(command Command, payload []byte) (DispatchReply, error) {
	switch command {
	case LOAD:
		logger.Debug("Welcome received command LOAD")
		sf, err := session.LoadSplitFile(machine.persister)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to load split file from Welcome state: %s", err))
			return DispatchReply{}, err
		}
		machine.sessionService.SetLoadedSplitFile(sf)
		machine.changeState(RUNNING, machine.sessionService)
		return DispatchReply{}, nil
	case NEW:
		logger.Debug("Welcome received command NEW")
		machine.changeState(NEWFILE)
		return DispatchReply{}, nil
	default:
		return DispatchReply{}, fmt.Errorf("invalid command %d for state Welcome", command)
	}
}

// NewFile indicates that the frontend should show the SplitEditor, and cannot send along a split file.
//
// CANCEL from NewFile should return to the Welcome state
type NewFile struct{}

func NewNewFileState() (*NewFile, error) {
	return &NewFile{}, nil
}

func (n *NewFile) String() string {
	return "NewFile"
}

func (n *NewFile) OnEnter(ctx context.Context) error {
	runtime.EventsEmit(ctx, "state:enter", NEWFILE)
	return nil
}
func (n *NewFile) OnExit() error { return nil }
func (n *NewFile) Receive(command Command, payload []byte) (DispatchReply, error) {
	switch command {
	case CANCEL:
		machine.changeState(WELCOME, nil)
	case SUBMIT:
		var splitFilePayload = session.SplitFilePayload{}
		err := json.Unmarshal(payload, &splitFilePayload)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to unmarshal into split file from NewFile state: %s", err))
		} else {
			splitFile, err := session.UpdateSplitFile(nil, machine.persister, nil, splitFilePayload)
			machine.sessionService.SetLoadedSplitFile(splitFile)
			if err != nil {
				logger.Error(fmt.Sprintf("failed to update split file from NewFile state: %s", err))
			}
			machine.changeState(RUNNING, splitFile)
		}
	default:
		panic("unhandled default case")
	}
	return DispatchReply{}, nil
}

// Editing indicates that the frontend should show the SplitEditor, and can also pass along a split file if loaded.
type Editing struct {
	ctx context.Context
}

// NewEditingState requires a *session.SplitFile as its sole payload parameter
func NewEditingState() (*Editing, error) {
	return &Editing{}, nil
}

// OnEnter sets the context from Wails, and signals the frontend to show the SplitEditor with the specified split file (or nil)
func (e *Editing) OnEnter(ctx context.Context) error {
	e.ctx = ctx
	payload := machine.sessionService.GetLoadedSplitFile()
	machine.sessionService.Pause()
	runtime.EventsEmit(ctx, "state:enter", EDITING, payload)
	return nil
}

func (e *Editing) OnExit() error { return nil }
func (e *Editing) Receive(command Command, payload []byte) (DispatchReply, error) {
	switch command {
	case CANCEL:
		machine.changeState(RUNNING, nil)
	case SUBMIT:
		var splitFilePayload = session.SplitFilePayload{}
		err := json.Unmarshal(payload, &splitFilePayload)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to unmarshal into split file from Editing state: %s", err))
		}
		splitFile, err := session.UpdateSplitFile(machine.runtimeProvider, machine.persister, machine.sessionService.GetLoadedSplitFile(), splitFilePayload)
		machine.sessionService.SetLoadedSplitFile(splitFile)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to update split file from Editing state: %s", err))
		} else {
			machine.changeState(RUNNING, nil)
			return DispatchReply{}, nil
		}
	default:
		panic("unhandled default case")

	}
	return DispatchReply{}, nil
}

func (e *Editing) String() string {
	return "Editing"
}

// Running represents the state where a splitfile has been loaded, the UI should be showing the SplitList and the timer.
type Running struct {
	splitFile *session.SplitFile
}

func NewRunningState() (*Running, error) {
	return &Running{}, nil
}

func (r Running) OnEnter(ctx context.Context) error {
	payload := machine.sessionService.GetSessionStatus()
	runtime.EventsEmit(ctx, "state:enter", RUNNING, payload)
	logger.Debug(fmt.Sprintf("state:enter event emittted : %v (%T)", payload.SplitFile, payload))
	return nil
}

func (r Running) OnExit() error {
	return nil
}

func (r Running) Receive(command Command, params []byte) (DispatchReply, error) {
	switch command {
	case CLOSE:
		logger.Debug(fmt.Sprintf("Running received CLOSE command: %s", params))
		machine.sessionService.CloseSplitFile()
		machine.changeState(WELCOME, nil)
	case EDIT:
		logger.Debug(fmt.Sprintf("Running received EDIT command: %s", params))
		machine.changeState(EDITING, nil)
	case SAVE:
		logger.Debug(fmt.Sprintf("Running received SAVE command: %s", params))
		windowParams := session.WindowParams{}
		err := json.Unmarshal(params, &windowParams)
		if err != nil {
			msg := fmt.Sprintf("failed to unmarshal into window params: %s - %s", err, string(params))
			logger.Error(msg)
			return DispatchReply{Code: 1, Message: msg}, err
		}
		err = r.splitFile.Save(machine.persister, windowParams)
		if err != nil {
			msg := fmt.Sprintf("failed to save split file to session: %s", err)
			logger.Error(msg)
			return DispatchReply{Code: 2, Message: msg}, err
		}
	case SPLIT:
		logger.Debug(fmt.Sprintf("Running received SPLIT command: %s", params))
		machine.sessionService.Split()
		return DispatchReply{}, nil
	default:
		panic("unhandled default case in Running")
	}

	return DispatchReply{}, nil
}

func (r Running) String() string {
	return "Running"
}
