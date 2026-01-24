package dispatcher

import "testing"

type mockDispatchReceiver struct{}

func (r mockDispatchReceiver) ReceiveDispatch(Command, *string) (DispatchReply, error) {
	return DispatchReply{
		Code:    69,
		Message: "Nice.",
	}, nil
}

func TestDispatch(t *testing.T) {
	dr := mockDispatchReceiver{}
	s := NewService(dr, nil, "")
	reply, _ := s.Dispatch(SPLIT, nil)
	if reply.Code != 69 || reply.Message != "Nice." {
		t.Fatalf("Dispatch expected to return code 69 with message Nice. but got %v: %s", reply.Code, reply.Message)
	}
}
