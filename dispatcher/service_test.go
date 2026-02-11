package dispatcher

import (
	"testing"

	"github.com/zellydev-games/opensplit/command"
)

type mockDispatchReceiver struct{}

func (r mockDispatchReceiver) ReceiveDispatch(command.Command, *string) (DispatchReply, error) {
	return DispatchReply{
		Code:    69,
		Message: "Nice.",
	}, nil
}

func TestDispatch(t *testing.T) {
	dr := mockDispatchReceiver{}
	s := NewService(dr, nil, nil, nil)
	reply, _ := s.Dispatch(command.SPLIT, nil)
	if reply.Code != 69 || reply.Message != "Nice." {
		t.Fatalf("Dispatch expected to return code 69 with message Nice. but got %v: %s", reply.Code, reply.Message)
	}
}
