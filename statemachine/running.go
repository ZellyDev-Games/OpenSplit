package statemachine

import (
	"fmt"
	"strconv"
	"time"

	"github.com/zellydev-games/opensplit/bridge"
	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/repo/adapters"
	"github.com/zellydev-games/opensplit/session"
)

type Running struct{}

func NewRunningState() (*Running, error) {
	return &Running{}, nil
}

func (r *Running) OnEnter() error {
	machine.saveOnWindowDimensionChanges = true

	return nil
}

func (r *Running) EmitUI() error {
	bridge.EmitUIEvent(
		machine.runtimeProvider,
		bridge.AppViewModel{
			View:    bridge.AppViewRunning,
			Session: adapters.DomainToDTO(machine.sessionService),
			Config:  machine.configService,
		},
	)

	return nil
}

func (r *Running) OnExit() error {
	machine.saveOnWindowDimensionChanges = false

	return nil
}

func (r *Running) Receive(
	c command.Command,
	payload *string,
) (dispatcher.DispatchReply, error) {
	switch c {
	case command.CLOSE:
		logger.Debug(
			logModule,
			"Running received CLOSE c",
		)

		_, err := machine.promptDirtySave()
		if err != nil {
			return dispatcher.DispatchReply{}, err
		}

		if err := machine.skinProvider.SetSkin(
			machine.configService.SelectedSkin,
			false,
		); err != nil {
			logger.Errorf(
				logModule,
				"failed to set skin: %v",
				err,
			)
		}

		machine.sessionService.CloseRun()
		machine.repoService.Close()
		machine.changeState(WELCOME, nil)

	case command.EDIT:
		logger.Debug(
			logModule,
			"Running received EDIT c",
		)

		if _, ok := machine.sessionService.Run(); ok {
			return dispatcher.DispatchReply{
				Code:    1,
				Message: "can't edit splitfile mid run",
			}, nil
		}

		machine.changeState(EDITING, nil)

	case command.SAVE:
		logger.Debug(
			logModule,
			"Running received SAVE c",
		)

		err := machine.saveSplitFile()
		if err != nil {
			msg := fmt.Sprintf(
				"failed to save split file to session: %s",
				err,
			)

			logger.Error(
				logModule,
				msg,
			)

			return dispatcher.DispatchReply{
				Code:    2,
				Message: msg,
			}, err
		}

	case command.SPLIT:
		logger.Debug(
			logModule,
			"Running received SPLIT c",
		)

		r.restoreRaceDoneForProgress()
		result := machine.sessionService.Split()

		if result == session.SplitFinished {
			machine.notifyRunDone()
			machine.runtimeProvider.EventsEmit(
				"opensplit:done",
			)
		}

	case command.UNDO:
		logger.Debug(
			logModule,
			"Running received UNDO",
		)

		wasRaceDone := machine.sessionService.RaceDoneActive()
		if wasRaceDone {
			r.restoreRaceDoneForProgress()
		}
		wasFinished := machine.sessionService.State() == session.Finished
		machine.sessionService.Undo()
		afterUndo := machine.sessionService.State()
		if wasFinished && !wasRaceDone &&
			(afterUndo == session.Running || afterUndo == session.Paused) {
			machine.notifyRunUndone()
			machine.runtimeProvider.EventsEmit("opensplit:undone")
		}

	case command.SKIP:
		logger.Debug(
			logModule,
			"Running received SKIP",
		)

		r.restoreRaceDoneForProgress()
		machine.sessionService.Skip()

	case command.PAUSE:
		logger.Debug(
			logModule,
			"Running received PAUSE",
		)
		if payload != nil && *payload != "" {
			machine.sessionService.SetForfeitPaused(*payload == "1")
		} else {
			machine.sessionService.Pause()
		}

	case command.RESET:
		logger.Debug(
			logModule,
			"Running received RESET",
		)

		_ = machine.promptPartialRun()
		machine.sessionService.ClearRuntimeOffsetOverride()
		machine.sessionService.Reset()
		machine.notifyRunForfeit()

	case command.DONE:
		logger.Debug(
			logModule,
			"Running received DONE c",
		)
		machine.sessionService.RaceDone()

	case command.UNDONE:
		logger.Debug(
			logModule,
			"Running received UNDONE c",
		)

		machine.sessionService.RaceUndone()

	case command.SET_RUNTIME_OFFSET:
		if payload == nil {
			return dispatcher.DispatchReply{
				Code:    10,
				Message: "missing offset payload",
			}, nil
		}

		ms, err := strconv.ParseInt(
			*payload,
			10,
			64,
		)

		if err != nil {
			return dispatcher.DispatchReply{
				Code:    11,
				Message: "invalid offset payload",
			}, nil
		}

		logger.Infof(
			logModule,
			"runtime offset set to %dms",
			ms,
		)

		machine.sessionService.SetRuntimeOffsetOverride(
			time.Duration(ms) * time.Millisecond,
		)

	case command.CLEAR_RUNTIME_OFFSET:
		logger.Info(
			logModule,
			"runtime offset cleared",
		)
		machine.sessionService.ClearRuntimeOffsetOverride()

	case command.COMPARISON_LEFT:
		machine.runtimeProvider.EventsEmit(
			"comparison:left",
		)

	case command.COMPARISON_RIGHT:
		machine.runtimeProvider.EventsEmit(
			"comparison:right",
		)

	case command.TOGGLEWR:
		logger.Debug(
			logModule,
			"world record display toggled",
		)

		show, err :=
			machine.sessionService.ToggleWorldRecordDisplay()

		if err != nil {
			return dispatcher.DispatchReply{
				Code:    1,
				Message: err.Error(),
			}, nil
		}

		if err := machine.repoService.SaveWorldRecordDisplay(
			show,
		); err != nil {
			logger.Errorf(
				logModule,
				"failed saving WR display state: %v",
				err,
			)
		}

		machine.runtimeProvider.EventsEmit(
			"session:update",
			adapters.DomainToDTO(machine.sessionService),
		)

	case command.SETLAYOUT:
		if payload == nil {
			return dispatcher.DispatchReply{
				Code:    1,
				Message: "missing layout payload",
			}, nil
		}

		layout := *payload

		if layout != "vertical" &&
			layout != "horizontal" {
			return dispatcher.DispatchReply{
				Code:    2,
				Message: "invalid layout payload",
			}, nil
		}

		logger.Debugf(
			logModule,
			"splitter layout changed to %s",
			layout,
		)

		if err := machine.sessionService.SetLayout(
			layout,
		); err != nil {
			return dispatcher.DispatchReply{
				Code:    3,
				Message: err.Error(),
			}, nil
		}

		if err := machine.repoService.SaveSplitFileLayout(
			layout,
		); err != nil {
			logger.Errorf(
				logModule,
				"failed saving splitter layout: %v",
				err,
			)

			return dispatcher.DispatchReply{
				Code:    4,
				Message: err.Error(),
			}, nil
		}

		// Emit a complete UI model so App.tsx applies the window
		// dimensions associated with the newly selected layout.
		if err := r.EmitUI(); err != nil {
			logger.Errorf(
				logModule,
				"failed to emit UI after layout change: %v",
				err,
			)

			return dispatcher.DispatchReply{
				Code:    5,
				Message: err.Error(),
			}, nil
		}

	default:
		logger.Warnf(
			logModule,
			"unhandled default case in Running: %d",
			c,
		)
	}

	return dispatcher.DispatchReply{}, nil
}

func (r *Running) restoreRaceDoneForProgress() {
	if !machine.sessionService.RaceDoneActive() {
		return
	}
	machine.sessionService.RaceUndone()
	machine.notifyRunUndone()
	machine.runtimeProvider.EventsEmit("opensplit:undone")
}

func (r *Running) String() string {
	return "Running"
}

func (r *Running) ID() StateID {
	return RUNNING
}
