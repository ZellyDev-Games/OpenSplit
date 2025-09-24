package statemachine

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/session"
	"github.com/zellydev-games/opensplit/splitfile"
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
	OnEnter() error
	OnExit() error
	Receive(command Command, payload []byte) (DispatchReply, error)
	String() string
}

// Service represents a state machine and holds references to all the tools to allow states to do useful work
type Service struct {
	currentState    state
	sessionService  *session.Service
	runtimeProvider session.RuntimeProvider
}

// InitMachine sets the global singleton, and gives it a friendly default state
func InitMachine(runtimeProvider session.RuntimeProvider, sessionService *session.Service) *Service {
	machine = &Service{
		sessionService:  sessionService,
		runtimeProvider: runtimeProvider,
	}

	return machine
}

// Startup is called by Wails.Run to pass in a context to use against Wails.platform
func (s *Service) Startup() {
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
		s.runtimeProvider.Quit()
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
		err := s.currentState.OnEnter()
		if err != nil {
			logger.Error(fmt.Sprintf("OnEnter failed: %v", err))
		}
	}
}

// Welcome greets the user by indicating the frontend should display the Welcome screen
type Welcome struct{}

// NewWelcomeState requires a *session.Service as its one and only payload parameter
func NewWelcomeState() (*Welcome, error) {
	return &Welcome{}, nil
}

// String returns the human friendly name for a State
func (w *Welcome) String() string {
	return "Welcome"
}

// OnEnter sets the context from the Wails app and signals the frontend to show the Welcome component
func (w *Welcome) OnEnter() error {
	machine.runtimeProvider.EventsEmit("state:enter", WELCOME)
	return nil
}
func (w *Welcome) OnExit() error { return nil }
func (w *Welcome) Receive(command Command, payload []byte) (DispatchReply, error) {
	switch command {
	case LOAD:
		logger.Debug("Welcome received command LOAD")
		err := machine.sessionService.Load()
		if err != nil {
			return DispatchReply{1, "failed to load splitfile: " + err.Error()}, err
		}
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

func (n *NewFile) OnEnter() error {
	machine.runtimeProvider.EventsEmit("state:enter", NEWFILE)
	return nil
}
func (n *NewFile) OnExit() error { return nil }
func (n *NewFile) Receive(command Command, payload []byte) (DispatchReply, error) {
	switch command {
	case CANCEL:
		machine.changeState(WELCOME, nil)
	case SUBMIT:
		err := machine.sessionService.SaveAs()
		if err != nil {
			return DispatchReply{1, "failed to save splitfile: " + err.Error()}, err
		}
		machine.changeState(RUNNING)
	default:
		panic("unhandled default case")
	}
	return DispatchReply{}, nil
}

// Editing indicates that the frontend should show the SplitEditor, and can also pass along a split file if loaded.
type Editing struct{}

// NewEditingState requires a *session.SplitFile as its sole payload parameter
func NewEditingState() (*Editing, error) {
	return &Editing{}, nil
}

// OnEnter sets the context from Wails, and signals the frontend to show the SplitEditor with the specified split file (or nil)
func (e *Editing) OnEnter() error {
	payload := machine.sessionService.SplitFile()
	machine.sessionService.Pause()
	machine.runtimeProvider.EventsEmit("state:enter", EDITING, payload)
	return nil
}

func (e *Editing) OnExit() error { return nil }
func (e *Editing) Receive(command Command, payload []byte) (DispatchReply, error) {
	switch command {
	case CANCEL:
		machine.changeState(RUNNING)
	case SUBMIT:
		err := machine.sessionService.Save()
		if err != nil {
			return DispatchReply{1, "failed to save splitfile: " + err.Error()}, err
		}
		machine.changeState(RUNNING)
		return DispatchReply{}, nil
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
	splitFile *splitfile.SplitFile
}

func NewRunningState() (*Running, error) {
	return &Running{}, nil
}

func (r Running) OnEnter() error {
	//payload := machine.sessionService.GetSessionStatus()
	//machine.runtimeProvider.EventsEmit("state:enter", RUNNING, payload)
	return nil
}

func (r Running) OnExit() error {
	return nil
}

func (r Running) Receive(command Command, params []byte) (DispatchReply, error) {
	switch command {
	case CLOSE:
		logger.Debug(fmt.Sprintf("Running received CLOSE command: %s", params))
		machine.sessionService.CloseRun()
		machine.changeState(WELCOME, nil)
	case EDIT:
		logger.Debug(fmt.Sprintf("Running received EDIT command: %s", params))
		machine.changeState(EDITING, nil)
	case SAVE:
		logger.Debug(fmt.Sprintf("Running received SAVE command: %s", params))
		windowParams := splitfile.WindowParams{}
		err := json.Unmarshal(params, &windowParams)
		if err != nil {
			msg := fmt.Sprintf("failed to unmarshal into window params: %s - %s", err, string(params))
			logger.Error(msg)
			return DispatchReply{Code: 1, Message: msg}, err
		}
		err = machine.sessionService.Save()
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
