package statemachine

import (
	"errors"
	"fmt"
	"sync"

	"github.com/zellydev-games/opensplit/bridge"
	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/keyinfo"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/repo/adapters"
)

const RecordingArmed = 10

// Config manages the configuration editing state and temporary hotkey
// recording.
type Config struct {
	mu             sync.Mutex
	listeningFor   command.Command
	recordingArmed bool
	previousState  StateID
}

func NewConfigState(previousState StateID) (*Config, error) {
	return &Config{
		previousState: previousState,
	}, nil
}

func (c *Config) OnEnter() error {
	logger.Debug(logModule, "entering config state")

	return nil
}

func (c *Config) EmitUI() error {
	bridge.EmitUIEvent(
		machine.runtimeProvider,
		bridge.AppViewModel{
			View:   bridge.AppViewSettings,
			Config: machine.configService,
		},
	)

	return nil
}

func (c *Config) OnExit() error {
	return nil
}

// Receive handles configuration commands originating from the frontend.
func (c *Config) Receive(cmd command.Command, payload *string) (dispatcher.DispatchReply, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch cmd {
	case command.SPLIT:
		fallthrough
	case command.UNDO:
		fallthrough
	case command.SKIP:
		fallthrough
	case command.PAUSE:
		fallthrough
	case command.COMPARISON_LEFT:
		fallthrough
	case command.COMPARISON_RIGHT:
		fallthrough
	case command.RESET:
		c.recordingArmed = true
		c.listeningFor = cmd
		logger.Infof(logModule, "recording armed for cmd: %d", c.listeningFor)
		err := machine.hotkeyProvider.StartHook(func(data keyinfo.KeyData) {
			c.handleHotkey(data)
			logger.Infof(logModule, "updated cmd %v with hotkey %s (%d)",
				c.listeningFor, data.LocaleName, data.KeyCode)
		})
		if err != nil {
			logger.Error(logModule, err.Error())
			c.recordingArmed = false
			return dispatcher.DispatchReply{Code: 6}, err
		}
		return dispatcher.DispatchReply{Code: RecordingArmed}, nil
	case command.CANCEL:
		machine.changeState(c.previousState)
		return dispatcher.DispatchReply{}, nil
	case command.SUBMIT:
		if payload == nil {
			return dispatcher.DispatchReply{
				Code:    4,
				Message: "missing config payload",
			}, errors.New("missing config payload")
		}

		newConfig, err := adapters.FrontEndToConfig([]byte(*payload))
		if err != nil {
			return dispatcher.DispatchReply{
				Code:    4,
				Message: err.Error(),
			}, err
		}

		machine.configService.Apply(newConfig)

		machine.configService.NotifyUpdate()

		if err := machine.repoService.SaveConfig(machine.configService); err != nil {
			message := fmt.Sprintf("error saving config to repo: %v", err)
			return dispatcher.DispatchReply{
				Code:    4,
				Message: message,
			}, errors.New(message)
		}

		machine.changeState(c.previousState)
		return dispatcher.DispatchReply{}, nil
	default:
		message := fmt.Sprintf("unknown cmd sent to config service: %v", cmd)
		return dispatcher.DispatchReply{Code: 5, Message: message}, errors.New(message)
	}
}

// handleHotkey stores the newly recorded hotkey binding.
func (c *Config) handleHotkey(data keyinfo.KeyData) {
	if c.recordingArmed {
		c.recordingArmed = false

		logger.Infof(logModule,
			"recording cmd %v -> keycode=%d mods=%v",
			c.listeningFor,
			data.KeyCode,
			data.Modifiers,
		)

		bridge.EmitHotkeyRecorded(
			machine.runtimeProvider,
			c.listeningFor,
			data,
		)
	}

	if err := machine.hotkeyProvider.Unhook(); err != nil {
		logger.Error(logModule, err.Error())
	}
}

func (c *Config) String() string {
	return "Config"
}

func (c *Config) ID() StateID {
	return CONFIG
}
