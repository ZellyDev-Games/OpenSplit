package statemachine

import (
	"fmt"

	"github.com/zellydev-games/opensplit/bridge"
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

			for command, keyData := range machine.configService.KeyConfig {
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
					_, _ = machine.ReceiveDispatch(command, nil)
					return
				} else {
					_, _ = machine.ReceiveDispatch(command, nil)
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

func (r *Running) Receive(command dispatcher.Command, _ *string) (dispatcher.DispatchReply, error) {
	switch command {
	case dispatcher.CLOSE:
		logger.Debug(logModule, "Running received CLOSE command")
		err := machine.promptDirtySave()
		if err != nil {
			return dispatcher.DispatchReply{}, err
		}
		machine.sessionService.CloseRun()
		machine.repoService.Close()
		machine.changeState(WELCOME, nil)
	case dispatcher.EDIT:
		logger.Debug(logModule, "Running received EDIT command")
		if _, ok := machine.sessionService.Run(); ok {
			return dispatcher.DispatchReply{Code: 1, Message: "can't edit splitfile mid run"}, nil
		}
		machine.changeState(EDITING, nil)
	case dispatcher.SAVE:
		logger.Debug(logModule, "Running received SAVE command")
		err := machine.saveSplitFile()
		if err != nil {
			msg := fmt.Sprintf("failed to save split file to session: %s", err)
			logger.Error(logModule, msg)
			return dispatcher.DispatchReply{Code: 2, Message: msg}, err
		}
	case dispatcher.SPLIT:
		logger.Debug(logModule, "Running received SPLIT command")
		machine.sessionService.Split()
	case dispatcher.UNDO:
		machine.sessionService.Undo()
	case dispatcher.SKIP:
		machine.sessionService.Skip()
	case dispatcher.PAUSE:
		machine.sessionService.Pause()
	case dispatcher.RESET:
		_ = machine.promptPartialRun()

		// note: promptPartialRun only adds the partial run to the session's loadedSplitFile's Runs slice.
		// Nothing has been saved to disk at this point, so keep the file dirty if needs be.
		machine.sessionService.Reset()
	default:
		logger.Warnf(logModule, "unhandled default case in Running: %d", command)
	}

	return dispatcher.DispatchReply{}, nil
}

func (r *Running) String() string {
	return "Running"
}
func (r *Running) ID() StateID {
	return RUNNING
}
