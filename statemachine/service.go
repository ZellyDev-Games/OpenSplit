package statemachine

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/zellydev-games/opensplit/bridge"
	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/config"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/keyinfo"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/repo"
	"github.com/zellydev-games/opensplit/repo/adapters"
	"github.com/zellydev-games/opensplit/session"
	"github.com/zellydev-games/opensplit/skin"
	"github.com/zellydev-games/opensplit/speedrun"
)

const logModule = "statemachine"

// machine is a private singleton instance of a *Service that represents
// the state machine.
var machine *Service

// StateID is a compact identifier for a State.
type StateID byte

const (
	WELCOME StateID = iota
	NEWFILE
	EDITING
	NEWSKIN
	EDITSKIN
	SKINEDITOR
	RUNNING
	CONFIG
)

// RuntimeProvider wraps Wails.runtimeProvider calls to allow for DI
// for testing.
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

type GlobalHotkeyProvider interface {
	Start(context.Context) error
	Configure(map[command.Command]keyinfo.KeyData) error
	Enable() error
	Disable() error
	SetCommandCallback(func(command.Command))
	Close() error
}

type uiEmitter interface {
	EmitUI() error
}

// state implementations can be operated by the Service and do meaningful work, and communicate state to the frontend
// via runtime.EventsEmit
type state interface {
	OnEnter() error
	OnExit() error
	Receive(c command.Command, payload *string) (dispatcher.DispatchReply, error)
	String() string
	ID() StateID
}

// Service represents a state machine and holds references to all the tools to allow states to do useful work
type Service struct {
	ctx                                   context.Context
	splitfileLock                         sync.Mutex
	currentState                          state
	sessionService                        *session.Service
	skinProvider                          skin.SkinProvider
	repoService                           *repo.Service
	runtimeProvider                       RuntimeProvider
	hotkeyProvider                        GlobalHotkeyProvider
	configService                         *config.Service
	speedrunService                       *speedrun.Service
	saveOnWindowDimensionChanges          bool
	unsubscribeFromWindowDimensionChanges func()
	windowHasFocus                        bool
	runDoneCallback                       func()
	runUndoneCallback                     func()
	runForfeitCallback                    func()
}

// SetRunEventCallbacks registers notifications for race actions driven by OpenSplit.
func (s *Service) SetRunEventCallbacks(done, undone, forfeit func()) {
	s.runDoneCallback = done
	s.runUndoneCallback = undone
	s.runForfeitCallback = forfeit
}

func (s *Service) notifyRunDone() {
	if s.runDoneCallback != nil {
		s.runDoneCallback()
	}
}

func (s *Service) notifyRunUndone() {
	if s.runUndoneCallback != nil {
		s.runUndoneCallback()
	}
}

func (s *Service) notifyRunForfeit() {
	if s.runForfeitCallback != nil {
		s.runForfeitCallback()
	}
}

func NewMachine(
	runtimeProvider RuntimeProvider,
	repoService *repo.Service,
	sessionService *session.Service,
	configService *config.Service,
	skinProvider skin.SkinProvider,
	speedrunService *speedrun.Service,
) *Service {
	machine = &Service{
		sessionService:  sessionService,
		runtimeProvider: runtimeProvider,
		repoService:     repoService,
		configService:   configService,
		skinProvider:    skinProvider,
		speedrunService: speedrunService,
	}

	return machine
}

func (s *Service) Startup(ctx context.Context) {
	logger.Info(
		logModule,
		"starting state machine",
	)

	s.ctx = ctx

	s.unsubscribeFromWindowDimensionChanges =
		s.setupWindowDimensionListener()

	logger.Debug(
		logModule,
		"registering ui:ready listener",
	)

	s.runtimeProvider.EventsOn(
		"ui:ready",
		func(...any) {
			logger.Debug(
				logModule,
				"ui:ready received",
			)

			if s.currentState == nil {
				logger.Error(
					logModule,
					"ui:ready received but currentState is nil",
				)
				return
			}

			logger.Debugf(
				logModule,
				"ui:ready current state=%s",
				s.currentState.String(),
			)

			if emitter, ok := s.currentState.(uiEmitter); ok {
				logger.Debug(
					logModule,
					"emitting UI model after ui:ready",
				)

				if err := emitter.EmitUI(); err != nil {
					logger.Errorf(
						logModule,
						"failed to emit UI after ui:ready: %v",
						err,
					)
				}
			} else {
				logger.Errorf(
					logModule,
					"current state %s does not implement uiEmitter",
					s.currentState.String(),
				)
			}
		},
	)

	logger.Debug(
		logModule,
		"initializing Welcome state",
	)

	s.changeState(WELCOME)

	logger.Debug(
		logModule,
		"state machine startup complete",
	)
}

func (s *Service) AttachHotkeyProvider(
	provider GlobalHotkeyProvider,
) {
	logger.Debug(
		logModule,
		"global hotkey provider attached",
	)

	s.hotkeyProvider = provider
}

func (s *Service) ReceiveDispatch(
	c command.Command,
	payload *string,
) (dispatcher.DispatchReply, error) {
	// FOCUS is a window/runtime event rather than a state-machine
	// command. It can arrive while the frontend is starting up,
	// before the initial state has been loaded.
	if c == command.FOCUS {
		if payload == nil {
			return dispatcher.DispatchReply{
				Code:    1,
				Message: `focus requires payload of "true" or "false"`,
			}, nil
		}

		s.windowHasFocus = *payload == "true"

		return dispatcher.DispatchReply{}, nil
	}

	// RaceTime controls can arrive before a split file is loaded. Keep the
	// runtime offset in the session so the next reset/start uses it.
	switch c {
	case command.SET_RUNTIME_OFFSET:
		if payload == nil {
			return dispatcher.DispatchReply{Code: 10, Message: "missing offset payload"}, nil
		}
		ms, err := strconv.ParseInt(*payload, 10, 64)
		if err != nil {
			return dispatcher.DispatchReply{Code: 11, Message: "invalid offset payload"}, nil
		}
		s.sessionService.SetRuntimeOffsetOverride(time.Duration(ms) * time.Millisecond)
		return dispatcher.DispatchReply{}, nil

	case command.CLEAR_RUNTIME_OFFSET:
		s.sessionService.ClearRuntimeOffsetOverride()
		return dispatcher.DispatchReply{}, nil
	}

	if s.currentState == nil {
		logger.Error(
			logModule,
			"c sent to state machine without a loaded state",
		)

		return dispatcher.DispatchReply{},
			errors.New(
				"c sent to state machine without a loaded state",
			)
	}

	switch c {
	case command.QUIT:
		logger.Debug(
			logModule,
			"QUIT c dispatched from front end",
		)

		s.runtimeProvider.Quit()

		return dispatcher.DispatchReply{}, nil

	case command.HELLO:
		return dispatcher.DispatchReply{
			Code:    0,
			Message: "HELLO",
		}, nil

	case command.TOGGLEGLOBAL:
		logger.Debug(
			logModule,
			"TOGGLEGLOBAL command dispatched from frontend",
		)

		active := !s.configService.GlobalHotkeysActive

		if s.hotkeyProvider != nil {
			var err error

			if active {
				err = s.hotkeyProvider.Enable()
			} else {
				err = s.hotkeyProvider.Disable()
			}

			if err != nil {
				message := "failed to change global hotkey state: " + err.Error()

				logger.Error(logModule, message)

				return dispatcher.DispatchReply{
					Code:    1,
					Message: message,
				}, err
			}
		}

		s.configService.GlobalHotkeysActive = active

		if err := s.repoService.SaveConfig(s.configService); err != nil {
			message := "failed to save global hotkey configuration: " + err.Error()

			logger.Error(logModule, message)

			return dispatcher.DispatchReply{
				Code:    2,
				Message: message,
			}, err
		}

		s.runtimeProvider.EventsEmit(
			"hotkeys:global-state",
			active,
		)

		return dispatcher.DispatchReply{
			Code:    0,
			Message: strconv.FormatBool(active),
		}, nil
	}

	logger.Debugf(
		logModule,
		"c %d dispatched to state %s",
		c,
		s.currentState.String(),
	)

	return s.currentState.Receive(
		c,
		payload,
	)
}

func (s *Service) changeState(
	newState StateID,
	args ...interface{},
) {
	if s.currentState != nil {
		logger.Debugf(
			logModule,
			"exiting state %s",
			s.currentState.String(),
		)

		err := s.currentState.OnExit()
		if err != nil {
			logger.Errorf(
				logModule,
				"OnExit failed: %v",
				err,
			)
		}
	}
	switch newState {
	case WELCOME:
		logger.Debug(
			logModule,
			"entering state Welcome",
		)

		s.currentState, _ = NewWelcomeState()

	case NEWFILE:
		logger.Debug(
			logModule,
			"entering state NewFile",
		)

		s.currentState, _ = NewNewFileState()

	case EDITING:
		logger.Debug(
			logModule,
			"entering state Editing",
		)

		s.currentState, _ = NewEditingState()

	case NEWSKIN:
		logger.Debug(
			logModule,
			"entering state NewSkin",
		)

		s.currentState, _ = NewNewSkinState()

	case EDITSKIN:
		logger.Debug(
			logModule,
			"entering state EditSkin",
		)

		s.currentState, _ = NewEditSkinState()

	case SKINEDITOR:
		logger.Debug(
			logModule,
			"entering skin editor",
		)

		layout := "vertical"

		if len(args) > 0 {
			if value, ok := args[0].(string); ok {
				if value == "horizontal" || value == "vertical" {
					layout = value
				}
			}
		}

		s.currentState, _ = NewSkinEditorState(layout)

	case RUNNING:
		logger.Debug(
			logModule,
			"entering state Running",
		)

		s.currentState, _ = NewRunningState()

	case CONFIG:
		logger.Debug(
			logModule,
			"entering state Config",
		)

		s.currentState, _ =
			NewConfigState(s.currentState.ID())

	default:
		panic("unhandled default case")
	}

	if s.currentState != nil {
		err := s.currentState.OnEnter()

		if err != nil {
			logger.Errorf(
				logModule,
				"OnEnter failed: %v",
				err,
			)
		}

		if emitter, ok := s.currentState.(uiEmitter); ok {
			if err := emitter.EmitUI(); err != nil {
				logger.Errorf(
					logModule,
					"EmitUI failed: %v",
					err,
				)
			}
		}
	}
}

func (s *Service) updateWorldRecord() {
	logger.Debug(
		logModule,
		"Updating World Record",
	)

	sf, ok := s.sessionService.SplitFile()
	if !ok {
		return
	}

	logger.Debugf(
		logModule,
		"GameID=%q CategoryID=%q",
		sf.GameID,
		sf.CategoryID,
	)

	logger.Debugf(
		logModule,
		"loaded splitfile: game=%q category=%q",
		sf.GameID,
		sf.CategoryID,
	)

	logger.Debug(
		logModule,
		"checking categoryID",
	)

	if sf.CategoryID == "" {
		return
	}

	logger.Debug(
		logModule,
		"Searching for New World Record",
	)

	variables := make([]speedrun.WRVariable, 0, len(sf.Variables))
	for _, variable := range sf.Variables {
		variables = append(variables, speedrun.WRVariable{ID: variable.ID, ValueID: variable.ValueID})
	}
	wr, err := s.speedrunService.SearchWR(sf.GameID, sf.CategoryID, variables)
	if err != nil {
		logger.Error(
			logModule,
			err.Error(),
		)

		return
	}

	logger.Infof(
		logModule,
		"loaded WR for category %s",
		sf.CategoryID,
	)

	showWorldRecord := sf.WR.Show
	sf.WR =
		s.speedrunService.ToWorldRecord(wr)

	sf.WR.Show = showWorldRecord

	logger.Debugf(
		logModule,
		"world record display=%v",
		sf.WR.Show,
	)

	s.sessionService.SetLoadedSplitFile(sf)

	logger.Debug(
		logModule,
		"Emiting New World Record",
	)

	bridge.EmitUIEvent(
		s.runtimeProvider,
		bridge.AppViewModel{
			View:    bridge.AppViewRunning,
			Session: adapters.DomainToDTO(s.sessionService),
			Config:  s.configService,
		},
	)
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

	logger.Debug(
		logModule,
		"saving split file",
	)

	err := s.repoService.SaveSplitFile(dto)
	if err != nil {
		return err
	}

	logger.Info(
		logModule,
		"split file saved",
	)

	s.sessionService.ClearDirty()

	return nil
}

func (s *Service) setupWindowDimensionListener() func() {
	return s.runtimeProvider.EventsOn(
		"window:dimensions",
		func(data ...any) {
			if !s.saveOnWindowDimensionChanges {
				return
			}

			logger.Infof(
				logModule,
				"Window dimensions have changed: x:%f y:%f w:%f h:%f",
				data...,
			)

			x := 10
			y := 10
			w := 100
			h := 100

			if len(data) > 0 {
				if f, ok := data[0].(float64); ok {
					x = max(10, int(f))
				}
			}

			if len(data) > 1 {
				if f, ok := data[1].(float64); ok {
					y = max(10, int(f))
				}
			}

			if len(data) > 2 {
				if f, ok := data[2].(float64); ok {
					w = max(100, int(f))
				}
			}

			if len(data) > 3 {
				if f, ok := data[3].(float64); ok {
					h = max(100, int(f))
				}
			}

			if err := s.repoService.SaveSplitFileWindowDimensions(
				x,
				y,
				w,
				h,
			); err != nil {
				logger.Errorf(
					logModule,
					"SaveSplitFileWindowDimensions failed: %v",
					err,
				)
			}

			s.sessionService.UpdateWindowDimensions(
				x,
				y,
				w,
				h,
			)
		},
	)
}

func (s *Service) promptPartialRun() error {
	run, ok := s.sessionService.Run()

	if ok && !run.Completed {
		response, err :=
			s.runtimeProvider.MessageDialog(
				runtime.MessageDialogOptions{
					Type:          runtime.QuestionDialog,
					Title:         "Add partial run splits to session?",
					Message:       "Do you want to save the splits from this partial run?",
					Buttons:       []string{"Yes", "No"},
					DefaultButton: "Yes",
				},
			)

		if err != nil {
			return err
		}

		if response == "Yes" {
			logger.Info(
				logModule,
				"persisting partial run",
			)

			s.sessionService.PersistRunToSession()

			return nil
		}
	}

	logger.Debug(
		logModule,
		"discarding partial run",
	)

	return nil
}
func (s *Service) promptDirtySave() (bool, error) {
	if !s.sessionService.Dirty() {
		return true, nil
	}

	response, err :=
		s.runtimeProvider.MessageDialog(
			runtime.MessageDialogOptions{
				Type:          runtime.QuestionDialog,
				Title:         "Save Changes?",
				Message:       "You have unsaved runs. Would you like to save them before closing?",
				Buttons:       []string{"Yes", "No"},
				DefaultButton: "Yes",
			},
		)

	if err != nil {
		return false, err
	}

	switch response {
	case "Yes":
		logger.Info(
			logModule,
			"saving unsaved runs before close",
		)

		if err := s.saveSplitFile(); err != nil {
			return false, err
		}

		return true, nil

	case "No":
		logger.Info(
			logModule,
			"discarding unsaved runs",
		)

		return true, nil
	}

	return false, nil
}
