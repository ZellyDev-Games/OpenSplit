package statemachine

import (
	"errors"
	"fmt"
	"sync"

	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/repo/adapters"
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
	machine.runtimeProvider.EventsEmit("state:enter", CONFIG, machine.configService)
	return nil
}

func (c *Config) OnExit() error {
	return nil
}

func (c *Config) Receive(command dispatcher.Command, payload *string) (dispatcher.DispatchReply, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.recordingArmed {
		if command == dispatcher.CANCEL {
			c.recordingArmed = false
			return dispatcher.DispatchReply{}, nil
		}

		if command != dispatcher.SUBMIT {
			message := fmt.Sprintf("wanted command SUBMIT while listening for new hotkey, got %v", command)
			c.recordingArmed = false
			return dispatcher.DispatchReply{Code: 1, Message: message}, errors.New(message)
		}

		if payload == nil {
			message := "submit payload while listening for hotkey was nil"
			c.recordingArmed = false
			return dispatcher.DispatchReply{Code: 2, Message: message}, errors.New(message)
		}

		info, err := adapters.PayloadToConfigKeyInfo([]byte(*payload))
		if err != nil {
			message := fmt.Sprintf("error parsing payload into KeyInfo %s", err)
			c.recordingArmed = false
			return dispatcher.DispatchReply{Code: 3, Message: message}, errors.New(message)
		}
		machine.hotkeyProvider.StopDispatcher()
		machine.configService.UpdateKeyBinding(c.listeningFor, info)
		message := fmt.Sprintf("updated command %v with hotkey %s (%d)", c.listeningFor, info.LocaleName, info.KeyCode)
		reply := dispatcher.DispatchReply{Message: message}
		logger.Info(message)
		c.recordingArmed = false
		return reply, nil
	} else {
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
			machine.hotkeyProvider.StartDispatcher(true)
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
			return dispatcher.DispatchReply{}, nil
		default:
			message := fmt.Sprintf("unknown command sent to config service: %v", command)
			return dispatcher.DispatchReply{Code: 5, Message: message}, errors.New(message)
		}
	}
}

func (c *Config) String() string {
	return "Config"
}

func (c *Config) ID() StateID {
	return CONFIG
}
