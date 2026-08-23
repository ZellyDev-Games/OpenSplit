package command

//go:generate go run tools/gencommands/main.go

// Command bytes are sent to dispatcher.Service.Dispatch.
// The active state decides how each command is handled.
type Command byte

const (
	QUIT Command = 0
)
