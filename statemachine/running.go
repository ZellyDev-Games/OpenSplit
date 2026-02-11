package statemachine

import (
	"fmt"

	"github.com/zellydev-games/opensplit/bridge"
	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/keyinfo"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/repo/adapters"
)

// Running represents the state where a dto has been loaded, the UI should be showing the SplitList and the timer.
type Running struct{}

func NewRunningState() (*Running, error) {
	return &Running{}, nil
}

func (r *Running) OnEnter() error {
	machine.saveOnWindowDimensionChanges = true
	sessionDto := adapters.DomainToDTO(machine.sessionService)
	if machine.hotkeyProvider != nil {
		err := machine.hotkeyProvider.StartHook(func(data keyinfo.KeyData) {
			if !machine.configService.GlobalHotkeysActive && !machine.windowHasFocus {
				return
			}

			for c, keyData := range machine.configService.KeyConfig {
				if keyData.KeyCode != data.KeyCode {
					continue
				}

				if len(keyData.Modifiers) != len(data.Modifiers) {
					continue
				}

				if len(keyData.Modifiers) > 0 {
					// Build lookup of pressed modifiers
					sent := make(map[int]struct{}, len(data.Modifiers))
					for _, m := range data.Modifiers {
						sent[m] = struct{}{}
					}

					// Ensure every required modifier exists
					match := true
					for _, required := range keyData.Modifiers {
						if _, ok := sent[required]; !ok {
							match = false
							break
						}
					}

					if !match {
						continue
					}
					_, _ = machine.ReceiveDispatch(c, nil)
					return
				} else {
					_, _ = machine.ReceiveDispatch(c, nil)
					return
				}
			}
		})

		if err != nil {
			logger.Error(logModule, err.Error())
			return err
		}
	}

	bridge.EmitUIEvent(machine.runtimeProvider, bridge.AppViewModel{
		View:    bridge.AppViewRunning,
		Session: sessionDto,
		Config:  machine.configService,
	})
	return nil
}

func (r *Running) OnExit() error {
	machine.saveOnWindowDimensionChanges = false
	if machine.hotkeyProvider != nil {
		err := machine.hotkeyProvider.Unhook()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Running) Receive(c command.Command, _ *string) (dispatcher.DispatchReply, error) {
	switch c {
	case command.CLOSE:
		logger.Debug(logModule, "Running received CLOSE c")
		err := machine.promptDirtySave()
		if err != nil {
			return dispatcher.DispatchReply{}, err
		}
		machine.sessionService.CloseRun()
		machine.repoService.Close()
		machine.changeState(WELCOME, nil)
	case command.EDIT:
		logger.Debug(logModule, "Running received EDIT c")
		if _, ok := machine.sessionService.Run(); ok {
			return dispatcher.DispatchReply{Code: 1, Message: "can't edit splitfile mid run"}, nil
		}
		machine.changeState(EDITING, nil)
	case command.SAVE:
		logger.Debug(logModule, "Running received SAVE c")
		err := machine.saveSplitFile()
		if err != nil {
			msg := fmt.Sprintf("failed to save split file to session: %s", err)
			logger.Error(logModule, msg)
			return dispatcher.DispatchReply{Code: 2, Message: msg}, err
		}
	case command.SPLIT:
		logger.Debug(logModule, "Running received SPLIT c")
		machine.sessionService.Split()
	case command.UNDO:
		machine.sessionService.Undo()
	case command.SKIP:
		machine.sessionService.Skip()
	case command.PAUSE:
		machine.sessionService.Pause()
	case command.RESET:
		_ = machine.promptPartialRun()

		// note: promptPartialRun only adds the partial run to the session's loadedSplitFile's Runs slice.
		// Nothing has been saved to disk at this point, so keep the file dirty if needs be.
		machine.sessionService.Reset()
	default:
		logger.Warnf(logModule, "unhandled default case in Running: %d", c)
	}

	return dispatcher.DispatchReply{}, nil
}

func (r *Running) String() string {
	return "Running"
}
func (r *Running) ID() StateID {
	return RUNNING
}
