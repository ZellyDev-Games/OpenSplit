package dispatcher

import (
	"sync"
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
	mu       sync.Mutex
	receiver DispatchReceiver
}

func NewService(receiver DispatchReceiver) *Service {
	return &Service{receiver: receiver}
}

func (s *Service) Dispatch(command Command, payload *string) (DispatchReply, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.receiver.ReceiveDispatch(command, payload)
}
