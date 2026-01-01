package statemachine

import (
	"errors"
	"fmt"
	"sync"

	"github.com/zellydev-games/opensplit/bridge"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/keyinfo"
	"github.com/zellydev-games/opensplit/logger"
)

const RecordingArmed = 10

type Config struct {
	mu             sync.Mutex
	listeningFor   dispatcher.Command
	recordingArmed bool
	previousState  StateID
}

func NewConfigState(previousState StateID) (*Config, error) {
	return &Config{
		previousState: previousState,
	}, nil
}

func (c *Config) OnEnter() error {
	bridge.EmitUIEvent(machine.runtimeProvider, bridge.AppViewModel{
		View:   bridge.AppViewSettings,
		Config: machine.configService,
	})
	return nil
}

func (c *Config) OnExit() error {
	return nil
}

func (c *Config) Receive(command dispatcher.Command, payload *string) (dispatcher.DispatchReply, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch command {
	case dispatcher.SPLIT:
		fallthrough
	case dispatcher.UNDO:
		fallthrough
	case dispatcher.SKIP:
		fallthrough
	case dispatcher.PAUSE:
		fallthrough
	case dispatcher.RESET:
		c.recordingArmed = true
		c.listeningFor = command
		logger.Info(fmt.Sprintf("recording armed for command: %d", c.listeningFor))
		err := machine.hotkeyProvider.StartHook(func(data keyinfo.KeyData) {
			c.handleHotkey(data)
			c.recordingArmed = false
			logger.Info(fmt.Sprintf("updated command %v with hotkey %s (%d)", c.listeningFor, data.LocaleName, data.KeyCode))
			err := machine.hotkeyProvider.Unhook()
			if err != nil {
				logger.Error(err.Error())
			}
		})
		if err != nil {
			logger.Error(err.Error())
			c.recordingArmed = false
			return dispatcher.DispatchReply{Code: 6}, err
		}
		return dispatcher.DispatchReply{Code: RecordingArmed}, nil
	case dispatcher.CANCEL:
		machine.changeState(c.previousState)
		return dispatcher.DispatchReply{}, nil
	case dispatcher.SUBMIT:
		err := machine.repoService.SaveConfig(machine.configService)
		if err != nil {
			message := fmt.Sprintf("error saving config to repo %s", err)
			return dispatcher.DispatchReply{Code: 4, Message: message}, errors.New(message)
		}

		machine.changeState(c.previousState)
		return dispatcher.DispatchReply{}, nil
	default:
		message := fmt.Sprintf("unknown command sent to config service: %v", command)
		return dispatcher.DispatchReply{Code: 5, Message: message}, errors.New(message)
	}
}

func (c *Config) handleHotkey(data keyinfo.KeyData) {
	if c.recordingArmed {
		c.recordingArmed = false
		machine.configService.UpdateKeyBinding(c.listeningFor, data)
	}
}

func (c *Config) String() string {
	return "Config"
}

func (c *Config) ID() StateID {
	return CONFIG
}
