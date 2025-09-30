package statemachine

import "github.com/zellydev-games/opensplit/dispatcher"

type Config struct {
	listeningFor string
}

func NewConfigState() (*Config, error) {
	return &Config{}, nil
}

func (c Config) OnEnter() error {
	return nil
}

func (c Config) OnExit() error {
	return nil
}

func (c Config) Receive(command dispatcher.Command, payload *string) (dispatcher.DispatchReply, error) {
	if c.listeningFor != "" {

	}
	hotkeyCommand := ""
	switch command {
	case dispatcher.SPLIT:
		hotkeyCommand = "SPLIT"
		goto gotCommand
	case dispatcher.UNDO:
		hotkeyCommand = "UNDO"
		goto gotCommand
	case dispatcher.SKIP:
		hotkeyCommand = "SKIP"
		goto gotCommand
	}

gotCommand:
	if hotkeyCommand != "" {
		//machine.configService.UpdateKeyBinding(hotkeyCommand)
	}
	return dispatcher.DispatchReply{}, nil
}

func (c Config) String() string {
	return "Config"
}
