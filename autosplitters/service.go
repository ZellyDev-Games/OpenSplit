package autosplitters

// TODO:
// check status of splits file
// update object variables
// add way to cancel autosplitting

import (
	"fmt"
	"time"

	nwa "github.com/zellydev-games/opensplit/autosplitters/NWA"
	qusb2snes "github.com/zellydev-games/opensplit/autosplitters/QUSB2SNES"
	"github.com/zellydev-games/opensplit/dispatcher"
)

type Splitters struct {
	NWAAutoSplitter       *nwa.NWASplitter
	QUSB2SNESAutoSplitter *qusb2snes.SyncClient
	UseAutosplitter       bool
	ResetTimerOnGameReset bool
	ResetGameOnTimerReset bool
	Addr                  string
	Port                  uint32
	Type                  AutosplitterType
}

type AutosplitterType int

const (
	NWA AutosplitterType = iota
	QUSB2SNES
)

func (s Splitters) Run(commandDispatcher *dispatcher.Service) {
	go func() {
		// loop trying to connect
		for {
			mil := 1 * time.Millisecond

			//check for split file loaded
			// if !splitsFile.loaded {
			// continue
			// }

			connectStart := time.Now()

			s.NWAAutoSplitter, s.QUSB2SNESAutoSplitter = s.newClient()

			if s.NWAAutoSplitter != nil || s.QUSB2SNESAutoSplitter != nil {

				if s.NWAAutoSplitter != nil {
					s.processNWA(commandDispatcher)
				}
				if s.QUSB2SNESAutoSplitter != nil {
					s.processQUSB2SNES(commandDispatcher)
				}
			}
			connectElapsed := time.Since(connectStart)
			// fmt.Println(mil - connectElapsed)
			time.Sleep(min(mil, max(0, mil-connectElapsed)))
		}
	}()
}

func (s Splitters) newClient() (*nwa.NWASplitter, *qusb2snes.SyncClient) {
	// fmt.Printf("Creating AutoSplitter Service\n")

	if s.UseAutosplitter {
		if s.Type == NWA {
			// fmt.Printf("Creating NWA AutoSplitter\n")
			client, connectError := nwa.Connect(s.Addr, s.Port)
			if connectError == nil {
				return &nwa.NWASplitter{
					Client: *client,
				}, nil
			} else {
				return nil, nil
			}
		}
		if s.Type == QUSB2SNES {
			// fmt.Printf("Creating QUSB2SNES AutoSplitter\n")
			client, connectError := qusb2snes.Connect(s.Addr, s.Port)
			if connectError == nil {
				return nil, client
			} else {
				return nil, nil
			}
		}
	}
	return nil, nil
}

// Memory should be moved out of here and received from the config file and sent to the splitter
func (s Splitters) processNWA(commandDispatcher *dispatcher.Service) {
	mil := 1 * time.Millisecond

	// // Battletoads test data
	// memData := []string{
	// 	("level,RAM,$0010,1")}
	// startConditionImport := []string{
	// 	("start:level,prior=0x0 && level,current=0x1")}
	// resetConditionImport := []string{
	// 	("reset:level,current=0x0 && level,prior≠0")}
	// splitConditionImport := []string{
	// 	// ("level2:level,prior=0xFF && level,current=0x2"),
	// 	("level3:level,prior=0xFF && level,current=0x3"),
	// 	("level4:level,prior=0xFF && level,current=0x4"),
	// 	// ("level5:level,prior=0xFF && level,current=0x5"),
	// 	("level6:level,prior=0xFF && level,current=0x6"),
	// 	// ("level7:level,prior=0xFF && level,current=0x7"),
	// 	("level8:level,prior=0xFF && level,current=0x8"),
	// 	("level9:level,prior=0xFF && level,current=0x9"),
	// 	("level10:level,prior=0xFF && level,current=0xA"),
	// 	("level11:level,prior=0xFF && level,current=0xB"),
	// 	("level12:level,prior=0xFF && level,current=0xC"),
	// 	("level13:level,prior=0xFF && level,current=0xD"),
	// }

	// Home Improvment test data
	memData := []string{
		("crates,WRAM,$001A8A,1"),
		("scene,WRAM,$00161F,1"),
		("W2P2HP,WRAM,$001499,1"),
		("W2P1HP,WRAM,$001493,1"),
		("BossHP,WRAM,$001491,1"),
		("state,WRAM,$001400,1"),
		("gameplay,WRAM,$000AE5,1"),
		("substage,WRAM,$000AE3,1"),
		("stage,WRAM,$000AE1,1"),
		("scene2,WRAM,$000886,1"),
		("play_state,WRAM,$0003B1,1"),
		("power_up,WRAM,$0003AF,1"),
		("weapon,WRAM,$0003CD,1"),
		("invul,WRAM,$001C05,1"),
		("FBossHP,WRAM,$00149D,1"),
	}

	startConditionImport := []string{
		("start:state,prior=0xC0 && state,current=0x0 && stage,current=0x0 && substage,current=0x0 && gameplay,current=0x11 && play_state,current=0x0"),
		("start:state,prior=0xD0 && state,current=0x0 && stage,current=0x0 && substage,current=0x0 && gameplay,current=0x11 && play_state,current=0x0"),
	}

	resetConditionImport := []string{
		("cutscene_reset:state,current=0x0 && state,prior=0xD0 && gameplay,prior=0x11 && gameplay,current=0x0"),
		("tool_reset:gameplay,prior=0x11 && gameplay,current=0x0 && scene,prior=0x4 && scene,current=0x0 && scene2,prior=0x3 && scene2,current=0x0"),
		("level_reset:gameplay,prior=0x13 && gameplay,current=0x0 && crates,current=0x0 && substage,current=0x0 && stage,current=0x0"),
	}

	splitConditionImport := []string{
		("1-1:state,prior=0xC8 && state,current=0x0 && stage,current=0x0 && substage,current=0x1 && gameplay,current=0x13"),
		("1-2:state,prior=0xC8 && state,current=0x0 && stage,current=0x0 && substage,current=0x2 && gameplay,current=0x13"),
		("1-3:state,prior=0xC8 && state,current=0x0 && stage,current=0x0 && substage,current=0x3 && gameplay,current=0x13"),
		("1-4:state,prior=0xC8 && state,current=0x0 && stage,current=0x0 && substage,current=0x4 && gameplay,current=0x13"),
		("1-5:state,prior=0xC8 && state,current=0x0 && stage,current=0x0 && substage,current=0x4 && gameplay,current=0x13 && BossHP,current=0x0"),
		("2-1:state,prior=0xC8 && state,current=0x0 && stage,current=0x1 && substage,current=0x1 && gameplay,current=0x13"),
		("2-2:state,prior=0xC8 && state,current=0x0 && stage,current=0x1 && substage,current=0x2 && gameplay,current=0x13"),
		("2-3:state,prior=0xC8 && state,current=0x0 && stage,current=0x1 && substage,current=0x3 && gameplay,current=0x13"),
		("2-4:state,prior=0xC8 && state,current=0x0 && stage,current=0x1 && substage,current=0x4 && gameplay,current=0x13"),
		("2-5:state,prior=0xC8 && state,current=0x0 && stage,current=0x1 && substage,current=0x4 && gameplay,current=0x13 && W2P1HP,current=0x0"),
		("3-1:state,prior=0xC8 && state,current=0x0 && stage,current=0x2 && substage,current=0x1 && gameplay,current=0x13"),
		("3-2:state,prior=0xC8 && state,current=0x0 && stage,current=0x2 && substage,current=0x2 && gameplay,current=0x13"),
		("3-3:state,prior=0xC8 && state,current=0x0 && stage,current=0x2 && substage,current=0x3 && gameplay,current=0x13"),
		("3-4:state,prior=0xC8 && state,current=0x0 && stage,current=0x2 && substage,current=0x4 && gameplay,current=0x13"),
		("3-5:state,prior=0xC8 && state,current=0x0 && stage,current=0x2 && substage,current=0x4 && gameplay,current=0x13 && BossHP,current=0x0"),
		("4-1:state,prior=0xC8 && state,current=0x0 && stage,current=0x3 && substage,current=0x1 && gameplay,current=0x13"),
		("4-2:state,prior=0xC8 && state,current=0x0 && stage,current=0x3 && substage,current=0x2 && gameplay,current=0x13"),
		("4-3:state,prior=0xC8 && state,current=0x0 && stage,current=0x3 && substage,current=0x3 && gameplay,current=0x13"),
		("4-4:state,prior=0xC8 && state,current=0x0 && stage,current=0x3 && substage,current=0x4 && gameplay,current=0x13"),
		("4-5:state,prior=0xC8 && state,current=0x0 && stage,current=0x3 && substage,current=0x4 && gameplay,current=0x13 && FBossHP,current=0xFF"),
	}

	s.NWAAutoSplitter.EmuInfo()      // Gets info about the emu; name, version, nwa_version, id, supported commands
	s.NWAAutoSplitter.EmuGameInfo()  // Gets info about the loaded game
	s.NWAAutoSplitter.EmuStatus()    // Gets the status of the emu
	s.NWAAutoSplitter.ClientID()     // Provides the client name to the NWA interface
	s.NWAAutoSplitter.CoreInfo()     // Might be useful to display the platform & core names
	s.NWAAutoSplitter.CoreMemories() // Get info about the memory banks available

	// receive setup data...probably through a channel
	//Setup Memory
	s.NWAAutoSplitter.MemAndConditionsSetup(memData, startConditionImport, resetConditionImport, splitConditionImport)

	splitCount := 0
	runStarted := false
	// this is the core loop of autosplitting
	// queries the device (emu, hardware, application) at the rate specified in ms
	for {
		processStart := time.Now()

		fmt.Printf("Checking for autosplitting updates.\n")
		autoState, err2 := s.NWAAutoSplitter.Update( /*TODO: Request Current Split*/ splitCount)
		if err2 != nil {
			return
		}
		if autoState.Start && !runStarted {
			//split run
			_, _ = commandDispatcher.Dispatch(dispatcher.SPLIT, nil)
			runStarted = !runStarted
		}
		if autoState.Split && runStarted {
			//split run
			_, _ = commandDispatcher.Dispatch(dispatcher.SPLIT, nil)
			splitCount++
		}
		if autoState.Reset && runStarted {
			if s.ResetTimerOnGameReset {
				_, _ = commandDispatcher.Dispatch(dispatcher.RESET, nil)
			}
			if s.ResetGameOnTimerReset {
				s.NWAAutoSplitter.SoftResetConsole()
			}
			splitCount = 0
			runStarted = !runStarted
		}
		// TODO: Close the connection after closing the splits file or receiving a disconnect signal
		// s.NWAAutoSplitter.Client.Close()

		processElapsed := time.Since(processStart)
		fmt.Println(processStart)
		fmt.Println(processElapsed)
		fmt.Println(mil - processElapsed)
		time.Sleep(min(mil, max(0, mil-processElapsed)))
	}
}

func (s Splitters) processQUSB2SNES(commandDispatcher *dispatcher.Service) {
	mil := 1 * time.Millisecond

	// Home Improvment test data
	memData := []string{
		("crates,0x1A8A,1"),
		("scene,0x161F,1"),
		("W2P2HP,0x1499,1"),
		("W2P1HP,0x1493,1"),
		("BossHP,0x1491,1"),
		("state,0x1400,1"),
		("gameplay,0x0AE5,1"),
		("substage,0x0AE3,1"),
		("stage,0x0AE1,1"),
		("scene2,0x0886,1"),
		("play_state,0x03B1,1"),
		("power_up,0x03AF,1"),
		("weapon,0x03CD,1"),
		("invul,0x1C05,1"),
		("FBossHP,0x149D,1"),
	}

	startConditionImport := []string{
		("start:state,prior=0xC0 && state,current=0x0 && stage,current=0x0 && substage,current=0x0 && gameplay,current=0x11 && play_state,current=0x0"),
		("start:state,prior=0xD0 && state,current=0x0 && stage,current=0x0 && substage,current=0x0 && gameplay,current=0x11 && play_state,current=0x0"),
	}

	resetConditionImport := []string{
		("cutscene_reset:state,current=0x0 && state,prior=0xD0 && gameplay,prior=0x11 && gameplay,current=0x0"),
		("tool_reset:gameplay,prior=0x11 && gameplay,current=0x0 && scene,prior=0x4 && scene,current=0x0 && scene2,prior=0x3 && scene2,current=0x0"),
		("level_reset:gameplay,prior=0x13 && gameplay,current=0x0 && crates,current=0x0 && substage,current=0x0 && stage,current=0x0"),
	}

	splitConditionImport := []string{
		("1-1:state,prior=0xC8 && state,current=0x0 && stage,current=0x0 && substage,current=0x1 && gameplay,current=0x13"),
		("1-2:state,prior=0xC8 && state,current=0x0 && stage,current=0x0 && substage,current=0x2 && gameplay,current=0x13"),
		("1-3:state,prior=0xC8 && state,current=0x0 && stage,current=0x0 && substage,current=0x3 && gameplay,current=0x13"),
		("1-4:state,prior=0xC8 && state,current=0x0 && stage,current=0x0 && substage,current=0x4 && gameplay,current=0x13"),
		("1-5:state,prior=0xC8 && state,current=0x0 && stage,current=0x0 && substage,current=0x4 && gameplay,current=0x13 && BossHP,current=0x0"),
		("2-1:state,prior=0xC8 && state,current=0x0 && stage,current=0x1 && substage,current=0x1 && gameplay,current=0x13"),
		("2-2:state,prior=0xC8 && state,current=0x0 && stage,current=0x1 && substage,current=0x2 && gameplay,current=0x13"),
		("2-3:state,prior=0xC8 && state,current=0x0 && stage,current=0x1 && substage,current=0x3 && gameplay,current=0x13"),
		("2-4:state,prior=0xC8 && state,current=0x0 && stage,current=0x1 && substage,current=0x4 && gameplay,current=0x13"),
		("2-5:state,prior=0xC8 && state,current=0x0 && stage,current=0x1 && substage,current=0x4 && gameplay,current=0x13 && W2P2HP,current=0x1 && W2P1HP,current=0x0 && BossHP,current=0x0"),
		("3-1:state,prior=0xC8 && state,current=0x0 && stage,current=0x2 && substage,current=0x1 && gameplay,current=0x13"),
		("3-2:state,prior=0xC8 && state,current=0x0 && stage,current=0x2 && substage,current=0x2 && gameplay,current=0x13"),
		("3-3:state,prior=0xC8 && state,current=0x0 && stage,current=0x2 && substage,current=0x3 && gameplay,current=0x13"),
		("3-4:state,prior=0xC8 && state,current=0x0 && stage,current=0x2 && substage,current=0x4 && gameplay,current=0x13"),
		("3-5:state,prior=0xC8 && state,current=0x0 && stage,current=0x2 && substage,current=0x4 && gameplay,current=0x13 && BossHP,current=0x0"),
		("4-1:state,prior=0xC8 && state,current=0x0 && stage,current=0x3 && substage,current=0x1 && gameplay,current=0x13"),
		("4-2:state,prior=0xC8 && state,current=0x0 && stage,current=0x3 && substage,current=0x2 && gameplay,current=0x13"),
		("4-3:state,prior=0xC8 && state,current=0x0 && stage,current=0x3 && substage,current=0x3 && gameplay,current=0x13"),
		("4-4:state,prior=0xC8 && state,current=0x0 && stage,current=0x3 && substage,current=0x4 && gameplay,current=0x13"),
		("4-5:state,prior=0xC8 && state,current=0x0 && stage,current=0x3 && substage,current=0x4 && gameplay,current=0x13 && FBossHP,current=0xFF"),
	}

	_ = s.QUSB2SNESAutoSplitter.SetName("OpenSplit")

	version, _ := s.QUSB2SNESAutoSplitter.AppVersion()
	fmt.Printf("Server version is %v\n", version)

	devices, _ := s.QUSB2SNESAutoSplitter.ListDevice()

	if len(devices) != 1 {
		if len(devices) == 0 {
			fmt.Printf("no devices present\n")
			return
		}
		fmt.Printf("unexpected devices: %#v\n", devices)
		return
	}
	device := devices[0]
	fmt.Printf("Using device %v\n", device)

	_ = s.QUSB2SNESAutoSplitter.Attach(device)
	fmt.Println("Connected.")

	info, _ := s.QUSB2SNESAutoSplitter.Info()
	fmt.Printf("%#v\n", info)

	var autosplitter = qusb2snes.NewQUSB2SNESAutoSplitter(memData, startConditionImport, resetConditionImport, splitConditionImport)

	splitCount := 0
	runStarted := false
	for {
		processStart := time.Now()

		summary, _ := autosplitter.Update(*s.QUSB2SNESAutoSplitter, splitCount)

		if summary.Start && !runStarted {
			_, _ = commandDispatcher.Dispatch(dispatcher.SPLIT, nil)
			runStarted = !runStarted
		}
		if summary.Split && runStarted {
			// IGT
			// timer.SetGameTime(*t)
			// RTA
			_, _ = commandDispatcher.Dispatch(dispatcher.SPLIT, nil)
			splitCount++
		}
		// need to get timer reset state
		if summary.Reset && runStarted /*|| timer is reset*/ {
			if s.ResetTimerOnGameReset {
				_, _ = commandDispatcher.Dispatch(dispatcher.RESET, nil)
			}
			if s.ResetGameOnTimerReset {
				_ = s.QUSB2SNESAutoSplitter.Reset()
			}
			autosplitter.ResetGameTracking()
			splitCount = 0
			runStarted = !runStarted
		}
		// TODO: Close the connection after closing the splits file or receiving a disconnect signal
		// s.QUSB2SNESAutoSplitter.Client.Close()

		processElapsed := time.Since(processStart)
		fmt.Println(processStart)
		fmt.Println(processElapsed)
		fmt.Println(mil - processElapsed)
		time.Sleep(min(mil, max(0, mil-processElapsed)))
	}
}
