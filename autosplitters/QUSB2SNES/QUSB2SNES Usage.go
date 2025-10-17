import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type LiveSplitCoreRenderer struct {
	appConfig *AppConfig
	timer     *Timer
	settings  *Settings
}

type AppConfig struct {
	mu                    sync.RWMutex
	GlobalHotkeys         *YesOrNo
	FrameRate             *float64
	PollingRate           *float64
	UseAutosplitter       *YesOrNo
	AutosplitterType      *AType
	ResetTimerOnGameReset *YesOrNo
	ResetGameOnTimerReset *YesOrNo
}

type YesOrNo int

const (
	No YesOrNo = iota
	Yes
)

type AType int

const (
	QUSB2SNES AType = iota
)

type ThreadEvent int

const (
	TimerReset ThreadEvent = iota
)

type Timer struct {
	mu sync.RWMutex
	// timer state fields here
}

func (t *Timer) Start() error {
	// start timer logic
	return nil
}

func (t *Timer) Reset(force bool) error {
	// reset timer logic
	return nil
}

func (t *Timer) SetGameTime(tSec float64) error {
	// set game time logic
	return nil
}

func (t *Timer) Split() error {
	// split timer logic
	return nil
}

type Settings struct{}

type AutoSplitter interface {
	Update(client *SyncClient) (*Summary, error)
	ResetGameTracking()
	GameTimeToSeconds() *float64
}

type SuperMetroidAutoSplitter struct {
	settings *Settings
}

func NewSuperMetroidAutoSplitter(settings *Settings) *SuperMetroidAutoSplitter {
	return &SuperMetroidAutoSplitter{settings: settings}
}

func (a *SuperMetroidAutoSplitter) Update(client *SyncClient) (*Summary, error) {
	// update logic
	return &Summary{}, nil
}

func (a *SuperMetroidAutoSplitter) ResetGameTracking() {
	// reset tracking logic
}

func (a *SuperMetroidAutoSplitter) GameTimeToSeconds() *float64 {
	// return game time in seconds
	return nil
}

type Summary struct {
	Start          bool
	Reset          bool
	Split          bool
	LatencyAverage float64
	LatencyStddev  float64
}

type SyncClient struct{}

func ConnectSyncClient() (*SyncClient, error) {
	// connect logic
	return &SyncClient{}, nil
}

func (c *SyncClient) SetName(name string) error {
	return nil
}

func (c *SyncClient) AppVersion() (string, error) {
	return "version", nil
}

func (c *SyncClient) ListDevice() ([]string, error) {
	return []string{"device1"}, nil
}

func (c *SyncClient) Attach(device string) error {
	return nil
}

func (c *SyncClient) Info() (interface{}, error) {
	return nil, nil
}

func (c *SyncClient) Reset() error {
	return nil
}

func messageBoxOnError(f func() error) {
	if err := f(); err != nil {
		fmt.Println("Error:", err)
	}
}

func (app *LiveSplitCoreRenderer) EnableGlobalHotkeys() error {
	// enable global hotkeys logic
	return nil
}

func appInit(
	app *LiveSplitCoreRenderer,
	syncReceiver <-chan ThreadEvent,
	cc *CreationContext,
) {
	context := cc.EguiCtx.Clone()
	context.SetVisuals(DarkVisuals())

	app.appConfig.mu.RLock()
	globalHotkeys := app.appConfig.GlobalHotkeys
	app.appConfig.mu.RUnlock()

	if globalHotkeys != nil && *globalHotkeys == Yes {
		messageBoxOnError(func() error {
			return app.EnableGlobalHotkeys()
		})
	}

	app.appConfig.mu.RLock()
	frameRate := DEFAULT_FRAME_RATE
	if app.appConfig.FrameRate != nil {
		frameRate = *app.appConfig.FrameRate
	}
	pollingRate := DEFAULT_POLLING_RATE
	if app.appConfig.PollingRate != nil {
		pollingRate = *app.appConfig.PollingRate
	}
	app.appConfig.mu.RUnlock()

	// Frame Rate Thread
	go func() {
		ticker := time.NewTicker(time.Duration(float64(time.Second) / frameRate))
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				context.Clone().RequestRepaint()
			}
		}
	}()

	timer := app.timer
	appConfig := app.appConfig

	appConfig.mu.RLock()
	useAutosplitter := appConfig.UseAutosplitter
	appConfig.mu.RUnlock()

	if useAutosplitter != nil && *useAutosplitter == Yes {
		appConfig.mu.RLock()
		autosplitterType := appConfig.AutosplitterType
		appConfig.mu.RUnlock()

		if autosplitterType != nil && *autosplitterType == QUSB2SNES {
			settings := app.settings

			go func() {
				for {
					latency := struct {
						mu    sync.RWMutex
						value [2]float64
					}{}

					err := func() error {
						client, err := ConnectSyncClient()
						if err != nil {
							return errors.Wrap(err, "creating usb2snes connection")
						}
						if err := client.SetName("annelid"); err != nil {
							return err
						}
						version, err := client.AppVersion()
						if err != nil {
							return err
						}
						fmt.Printf("Server version is %v\n", version)

						devices, err := client.ListDevice()
						if err != nil {
							return err
						}
						if len(devices) != 1 {
							if len(devices) == 0 {
								return errors.New("no devices present")
							}
							return errors.Errorf("unexpected devices: %#v", devices)
						}
						device := devices[0]
						fmt.Printf("Using device %v\n", device)

						if err := client.Attach(device); err != nil {
							return err
						}
						fmt.Println("Connected.")
						info, err := client.Info()
						if err != nil {
							return err
						}
						fmt.Printf("%#v\n", info)

						var autosplitter AutoSplitter = NewSuperMetroidAutoSplitter(settings)

						for {
							summary, err := autosplitter.Update(client)
							if err != nil {
								return err
							}
							if summary.Start {
								if err := timer.Start(); err != nil {
									return errors.Wrap(err, "start timer")
								}
							}
							if summary.Reset {
								appConfig.mu.RLock()
								resetTimerOnGameReset := appConfig.ResetTimerOnGameReset
								appConfig.mu.RUnlock()
								if resetTimerOnGameReset != nil && *resetTimerOnGameReset == Yes {
									if err := timer.Reset(true); err != nil {
										return errors.Wrap(err, "reset timer")
									}
								}
							}
							if summary.Split {
								if t := autosplitter.GameTimeToSeconds(); t != nil {
									if err := timer.SetGameTime(*t); err != nil {
										return errors.Wrap(err, "set game time")
									}
								}
								if err := timer.Split(); err != nil {
									return errors.Wrap(err, "split timer")
								}
							}

							latency.mu.Lock()
							latency.value[0] = summary.LatencyAverage
							latency.value[1] = summary.LatencyStddev
							latency.mu.Unlock()

							select {
							case ev := <-syncReceiver:
								if ev == TimerReset {
									autosplitter.ResetGameTracking()
									appConfig.mu.RLock()
									resetGameOnTimerReset := appConfig.ResetGameOnTimerReset
									appConfig.mu.RUnlock()
									if resetGameOnTimerReset != nil && *resetGameOnTimerReset == Yes {
										if err := client.Reset(); err != nil {
											return err
										}
									}
								}
							default:
							}

							time.Sleep(time.Duration(float64(time.Second) / pollingRate))
						}
					}()
					if err != nil {
						fmt.Println("Error:", err)
					}
				}
			}()

			time.Sleep(time.Second)
		}
	}
}

// Dummy types and functions to make the above compile

type CreationContext struct {
	EguiCtx *EguiContext
}

type EguiContext struct{}

func (e *EguiContext) Clone() *EguiContext {
	return &EguiContext{}
}

func (e *EguiContext) SetVisuals(v Visuals) {}

func (e *EguiContext) RequestRepaint() {}

type Visuals struct{}

func DarkVisuals() Visuals {
	return Visuals{}
}

const (
	DEFAULT_FRAME_RATE   = 60.0
	DEFAULT_POLLING_RATE = 30.0
)