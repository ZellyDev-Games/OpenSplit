package statemachine

import (
	"context"
	"errors"
	"fmt"
	"sync"

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

type HotkeyProvider interface {
	StartHook(func(data keyinfo.KeyData)) error
	Unhook() error
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
	hotkeyProvider                        HotkeyProvider
	configService                         *config.Service
	speedrunService                       *speedrun.Service
	saveOnWindowDimensionChanges          bool
	unsubscribeFromWindowDimensionChanges func()
	windowHasFocus                        bool
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

	s.changeState(WELCOME)

	s.runtimeProvider.EventsOn(
		"ui:ready",
		func(...any) {
			if emitter, ok := s.currentState.(uiEmitter); ok {
				_ = emitter.EmitUI()
			}
		},
	)
}

func (s *Service) AttachHotkeyProvider(
	provider HotkeyProvider,
) {
	logger.Debug(
		logModule,
		"hotkey provider attached",
	)

	s.hotkeyProvider = provider
}

func (s *Service) ReceiveDispatch(
	c command.Command,
	payload *string,
) (dispatcher.DispatchReply, error) {
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

	if c == command.QUIT {
		logger.Debug(
			logModule,
			"QUIT c dispatched from front end",
		)

		s.runtimeProvider.Quit()

		return dispatcher.DispatchReply{}, nil
	}

	if c == command.HELLO {
		return dispatcher.DispatchReply{
			Code:    0,
			Message: "HELLO",
		}, nil
	}

	if c == command.TOGGLEGLOBAL {
		logger.Debug(
			logModule,
			"TOGGLEGLOBAL c dispatched from frontend",
		)

		s.configService.GlobalHotkeysActive =
			!s.configService.GlobalHotkeysActive

		err := s.repoService.SaveConfig(
			s.configService,
		)

		if err != nil {
			message := fmt.Sprintf(
				"error saving config to repo %s",
				err,
			)

			return dispatcher.DispatchReply{
				Code:    1,
				Message: message,
			}, errors.New(message)
		}

		return dispatcher.DispatchReply{
			Message: fmt.Sprintf(
				"%t",
				s.configService.GlobalHotkeysActive,
			),
		}, nil
	}

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

	wr, err := s.speedrunService.SearchWR(sf.CategoryID)
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
