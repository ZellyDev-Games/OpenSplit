package statemachine

import (
	"errors"
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
	sessionDto := adapters.DomainToDTO(machine.sessionService)

	if machine.hotkeyProvider != nil {
		err := machine.hotkeyProvider.StartHook(func(data keyinfo.KeyData) {
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
			logger.Error(err.Error())
			return err
		}
	}

	bridge.EmitUIEvent(machine.runtimeProvider, bridge.AppViewModel{
		View:    bridge.AppViewRunning,
		Session: sessionDto,
	})
	return nil
}

func (r *Running) OnExit() error {
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
		logger.Debug("Running received CLOSE command")
		machine.sessionService.CloseRun()
		machine.repoService.Close()
		machine.changeState(WELCOME, nil)
	case dispatcher.EDIT:
		logger.Debug("Running received EDIT command")
		machine.changeState(EDITING, nil)
	case dispatcher.SAVE:
		logger.Debug("Running received SAVE command")
		sf, loaded := machine.sessionService.SplitFile()
		if !loaded {
			msg := "save called without loaded splitfile.  Wait, how did we get to running state without one?"
			return dispatcher.DispatchReply{Code: 3, Message: msg}, errors.New(msg)
		}
		w, h := machine.runtimeProvider.WindowGetSize()
		x, y := machine.runtimeProvider.WindowGetPosition()
		dto := adapters.DomainSplitFileToDTO(sf)
		err := machine.repoService.SaveSplitFile(dto, x, y, w, h)
		if err != nil {
			msg := fmt.Sprintf("failed to save split file to session: %s", err)
			logger.Error(msg)
			return dispatcher.DispatchReply{Code: 2, Message: msg}, err
		}
	case dispatcher.SPLIT:
		logger.Debug("Running received SPLIT command")
		machine.sessionService.Split()
	case dispatcher.UNDO:
		machine.sessionService.Undo()
	case dispatcher.SKIP:
		machine.sessionService.Skip()
	case dispatcher.PAUSE:
		machine.sessionService.Pause()
	case dispatcher.RESET:
		machine.sessionService.Reset()
	default:
		logger.Warn(fmt.Sprintf("unhandled default case in Running: %d", command))
	}

	return dispatcher.DispatchReply{}, nil
}

func (r *Running) String() string {
	return "Running"
}
func (r *Running) ID() StateID {
	return RUNNING
}
