package command

//go:generate go run tools/gencommands/main.go

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
	TOGGLEGLOBAL
	FOCUS
	TOGGLEWR
	HELLO
	DONE
	UNDONE
	SET_RUNTIME_OFFSET
	CLEAR_RUNTIME_OFFSET
	COMPARISON_LEFT
	COMPARISON_RIGHT
)
