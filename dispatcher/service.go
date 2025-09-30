package dispatcher

import (
	"encoding/json"
	"fmt"

	"github.com/labstack/gommon/log"
	"github.com/zellydev-games/opensplit/hotkeys"
	"github.com/zellydev-games/opensplit/logger"
)

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
	receiver DispatchReceiver
	keyMap   map[Command]hotkeys.KeyInfo
}

func NewService(receiver DispatchReceiver) *Service {
	return &Service{receiver: receiver}
}

func (s *Service) Dispatch(command Command, payload *string) (DispatchReply, error) {
	return s.receiver.ReceiveDispatch(command, payload)
}

// MapHotkey ranges through the loaded keyMap to find out if a button you just pressed is in it
//
// getRaw forces the mapper to Dispatch dispatcher.SUBMIT with the given hotkeys.KeyInfo as the payload instead of
// searching through the keyMap.
// This can be useful when you need raw key presses (e.g. when Config is listening to map the hotkeys to an action)
func (s *Service) MapHotkey(info hotkeys.KeyInfo, getRaw bool) error {
	if getRaw {
		infoBytes, err := json.Marshal(&info)
		infoString := string(infoBytes)
		_, err = s.Dispatch(SUBMIT, &infoString)
		return err
	}

	for command, keyInfo := range s.keyMap {
		if keyInfo.KeyCode == info.KeyCode {
			_, err := s.Dispatch(command, nil)
			if err != nil {
				log.Error(err)
			}
			return err
		}
	}
	logger.Debug(fmt.Sprintf("hotkey not found: %s (%d)", info.LocaleName, info.KeyCode))
	return nil
}
