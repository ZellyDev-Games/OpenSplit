type SupermetroidAutoSplitter struct {
	PriorState           uint8
	State                uint8
	PriorRoomID          uint16
	RoomID               uint16
	ResetTimerOnGameReset bool
	Client               NWASyncClient
}

type BattletoadsAutoSplitter struct {
	PriorLevel           uint8
	Level                uint8
	ResetTimerOnGameReset bool
	Client               NWASyncClient
}

type Splitter interface {
	ClientID()
	EmuInfo()
	EmuGameInfo()
	EmuStatus()
	CoreInfo()
	CoreMemories()
	Update() (NWASummary, error)
	Start() bool
	Reset() bool
	Split() bool
}

type Game int

const (
	Battletoads Game = iota
	SuperMetroid
)

func nwaobject(game Game, appConfig *sync.RWMutex, ip string, port uint32) Splitter {
	appConfig.RLock()
	defer appConfig.RUnlock()

	// Assuming appConfig is a struct pointer with ResetTimerOnGameReset field
	// This is a placeholder for actual config reading logic
	resetTimer := YesOrNo(0)
	// You need to implement actual reading from appConfig here

	switch game {
	case Battletoads:
		client, _ := (&NWASyncClient{}).Connect(ip, port)
		return &BattletoadsAutoSplitter{
			PriorLevel:           0,
			Level:                0,
			ResetTimerOnGameReset: resetTimer,
			Client:               *client,
		}
	case SuperMetroid:
		client, _ := (&NWASyncClient{}).Connect(ip, port)
		return &SupermetroidAutoSplitter{
			PriorState:           0,
			State:                0,
			PriorRoomID:          0,
			RoomID:               0,
			ResetTimerOnGameReset: resetTimer,
			Client:               *client,
		}
	default:
		return nil
	}
}

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/eventlog"
)

func appInit(
	app *LiveSplitCoreRenderer,
	syncReceiver <-chan ThreadEvent,
	cc *eframeCreationContext,
	appConfig *sync.RWMutex
) {
	context := cc.eguiCtx.Clone()
	context.SetVisuals(eguiVisualsDark())

	if app.appConfig.Read().GlobalHotkeys == YesOrNoYes {
		messageboxOnError(func() error {
			return app.enableGlobalHotkeys()
		})
	}

	frameRate := app.appConfig.Read().FrameRate
	if frameRate == 0 {
		frameRate = DefaultFrameRate
	}
	pollingRate := app.appConfig.Read().PollingRate
	if pollingRate == 0 {
		pollingRate = DefaultPollingRate
	}

	go func() {
		for {
			context.Clone().RequestRepaint()
			time.Sleep(time.Duration(1000.0/frameRate) * time.Millisecond)
		}
	}()

	timer := app.timer.Clone()
	appConfig := app.appConfig.Clone()

	if appConfig.Read().UseAutosplitter == YesOrNoYes {
		if appConfig.Read().AutosplitterType == autosplittersATypeNWA {
			game := app.game
			address := app.address.Read()
			port := *app.port.Read()

			go func() {
				for {
					client := nwaobject(game, appConfig, &address, port)
					err := printOnError(func() error {
						if err := client.emuInfo(); err != nil {
							return err
						}
						if err := client.emuGameInfo(); err != nil {
							return err
						}
						if err := client.emuStatus(); err != nil {
							return err
						}
						if err := client.clientID(); err != nil {
							return err
						}
						if err := client.coreInfo(); err != nil {
							return err
						}
						if err := client.coreMemories(); err != nil {
							return err
						}

						for {
							autoSplitStatus, err := client.update()
							if err != nil {
								return err
							}
							if autoSplitStatus.Start {
								if err := timer.WriteLock().Start(); err != nil {
									return errors.Wrap(err, "failed to start timer")
								}
							}
							if autoSplitStatus.Reset {
								if err := timer.WriteLock().Reset(true); err != nil {
									return errors.Wrap(err, "failed to reset timer")
								}
							}
							if autoSplitStatus.Split {
								if err := timer.WriteLock().Split(); err != nil {
									return errors.Wrap(err, "failed to split timer")
								}
							}

							time.Sleep(time.Duration(1000.0/pollingRate) * time.Millisecond)
						}
					})
					if err != nil {
						// handle error if needed
					}
					time.Sleep(1 * time.Second)
				}
			}()
		}
	}
}