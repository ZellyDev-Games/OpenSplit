package autosplitters

// TODO:
// check status of splits file
// update object variables
// qusb2snes usage
// load mem and condition data

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
	Addr                  string
	Port                  uint32
	Type                  AutosplitterType
}

type AutosplitterType int

const (
	NWA AutosplitterType = iota
	QUSB2SNES
)

// I don't think this should be here not sure why it's not updating the object
// func (s Splitters) Load() {
// useAuto := make(chan bool)
// resetTimer := make(chan bool)
// addr := make(chan string)
// port := make(chan uint32)
// aType := make(chan AutosplitterType)
// s.UseAutosplitter = <-useAuto
// s.ResetTimerOnGameReset = <-resetTimer
// s.Addr = <-addr
// s.Port = <-port
// s.Type = <-aType
// }

func (s Splitters) Run(commandDispatcher *dispatcher.Service) {
	go func() {
		// loop trying to connect
		for {
			mil := 2 * time.Millisecond

			//check for split file loaded
			// if !splitsFile.loaded {
			// continue
			// }

			connectStart := time.Now()

			s.NWAAutoSplitter, s.QUSB2SNESAutoSplitter = s.newClient( /*s.UseAutosplitter, s.ResetTimerOnGameReset, s.Addr, s.Port, s.Type*/ )

			if s.NWAAutoSplitter != nil || s.QUSB2SNESAutoSplitter != nil {

				if s.NWAAutoSplitter.Client.IsConnected() {
					s.processNWA(commandDispatcher)
				}
				// if s.QUSB2SNESAutoSplitter != nil {
				// s.processQUSB2SNES(commandDispatcher)
				// }
			}
			connectElapsed := time.Since(connectStart)
			// fmt.Println(mil - connectElapsed)
			time.Sleep(min(mil, max(0, mil-connectElapsed)))
		}
	}()
}

func (s Splitters) newClient( /*UseAutosplitter bool, ResetTimerOnGameReset bool, Addr string, Port uint32, Type AutosplitterType*/ ) (*nwa.NWASplitter, *qusb2snes.SyncClient) {
	// fmt.Printf("Creating AutoSplitter Service\n")

	if s.UseAutosplitter {
		if s.Type == NWA {
			// fmt.Printf("Creating NWA AutoSplitter\n")
			client, connectError := nwa.Connect(s.Addr, s.Port)
			if connectError == nil {
				return &nwa.NWASplitter{
					ResetTimerOnGameReset: s.ResetTimerOnGameReset,
					Client:                *client,
				}, nil
			} else {
				return nil, nil
			}
		}
		if s.Type == QUSB2SNES {
			fmt.Printf("Creating QUSB2SNES AutoSplitter\n")
			client, connectError := qusb2snes.Connect()
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
	mil := 2 * time.Millisecond

	// // Battletoads test data
	// memData := []string{
	//  ("level,RAM,$0010,1")}
	// // startConditionImport := []string{}
	// resetConditionImport := []string{
	//  ("reset:level,0,eeqc level,0,pnee")}
	// splitConditionImport := []string{
	//  ("start:level,0,peqe level,1,ceqe"),
	// 	("level:level,255,peqe level,2,ceqe"),
	// 	("level:level,255,peqe level,3,ceqe"),
	// 	("level:level,255,peqe level,4,ceqe"),
	// 	("level:level,255,peqe level,5,ceqe"),
	// 	("level:level,255,peqe level,6,ceqe"),
	// 	("level:level,255,peqe level,7,ceqe"),
	// 	("level:level,255,peqe level,8,ceqe"),
	// 	("level:level,255,peqe level,9,ceqe"),
	// 	("level:level,255,peqe level,10,ceqe"),
	// 	("level:level,255,peqe level,11,ceqe"),
	// 	("level:level,255,peqe level,12,ceqe"),
	// 	("level:level,255,peqe level,13,ceqe")}

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

	resetConditionImport := []string{
		("cutscene_reset:state,0,eeqc state,D0,peqe gameplay,11,peqe gameplay,0,eeqc"),
		("tool_reset:gameplay,11,peqe gameplay,0,eeqc scene,4,peqe scene,0,eeqc, scene2,3,peqe scene2,0,eeqc"),
		("level_reset:gameplay,13,peqe gameplay,0,eeqc crates,0,ceqe substage,0,ceqe stage,0,ceqe scene2,0,ceqe"),
	}

	splitConditionImport := []string{
		("start:state,C0,peqe state,0,ceqe stage,0,ceqe substage,0,ceqe gameplay,11,ceqe play_state,0,ceqe"),
		("start2:state,D0,peqe state,0,ceqe stage,0,ceqe substage,0,ceqe gameplay,11,ceqe play_state,0,ceqe"),
		("1-1:state,C8,peqe state,0,ceqe stage,0,ceqe substage,1,ceqe gameplay,13,ceqe"),
		("1-2:state,C8,peqe state,0,ceqe stage,0,ceqe substage,2,ceqe gameplay,13,ceqe"),
		("1-3:state,C8,peqe state,0,ceqe stage,0,ceqe substage,3,ceqe gameplay,13,ceqe"),
		("1-4:state,C8,peqe state,0,ceqe stage,0,ceqe substage,4,ceqe gameplay,13,ceqe"),
		("1-5:state,C8,peqe state,0,ceqe stage,0,ceqe substage,4,ceqe gameplay,13,ceqe BossHP,0,ceqe"),
		("2-1:state,C8,peqe state,0,ceqe stage,1,ceqe substage,1,ceqe gameplay,13,ceqe"),
		("2-2:state,C8,peqe state,0,ceqe stage,1,ceqe substage,2,ceqe gameplay,13,ceqe"),
		("2-3:state,C8,peqe state,0,ceqe stage,1,ceqe substage,3,ceqe gameplay,13,ceqe"),
		("2-4:state,C8,peqe state,0,ceqe stage,1,ceqe substage,4,ceqe gameplay,13,ceqe"),
		("2-5:state,C8,peqe state,0,ceqe stage,1,ceqe substage,4,ceqe gameplay,13,ceqe W2P2HP,1,ceqe W2P1HP,0,ceqe BossHP,0,ceqe"),
		("3-1:state,C8,peqe state,0,ceqe stage,2,ceqe substage,1,ceqe gameplay,13,ceqe"),
		("3-2:state,C8,peqe state,0,ceqe stage,2,ceqe substage,2,ceqe gameplay,13,ceqe"),
		("3-3:state,C8,peqe state,0,ceqe stage,2,ceqe substage,3,ceqe gameplay,13,ceqe"),
		("3-4:state,C8,peqe state,0,ceqe stage,2,ceqe substage,4,ceqe gameplay,13,ceqe"),
		("3-5:state,C8,peqe state,0,ceqe stage,2,ceqe substage,4,ceqe gameplay,13,ceqe BossHP,0,ceqe"),
		("4-1:state,C8,peqe state,0,ceqe stage,3,ceqe substage,1,ceqe gameplay,13,ceqe"),
		("4-2:state,C8,peqe state,0,ceqe stage,3,ceqe substage,2,ceqe gameplay,13,ceqe"),
		("4-3:state,C8,peqe state,0,ceqe stage,3,ceqe substage,3,ceqe gameplay,13,ceqe"),
		("4-4:state,C8,peqe state,0,ceqe stage,3,ceqe substage,4,ceqe gameplay,13,ceqe"),
		("4-5:state,C8,peqe state,0,ceqe stage,3,ceqe substage,4,ceqe gameplay,13,ceqe FBossHP,FF,ceqe"),
	}

	// receive setup data...probably through a channel
	//Setup Memory
	s.NWAAutoSplitter.MemAndConditionsSetup(memData /*startConditionImport,*/, resetConditionImport, splitConditionImport)

	s.NWAAutoSplitter.EmuInfo()      // Gets info about the emu; name, version, nwa_version, id, supported commands
	s.NWAAutoSplitter.EmuGameInfo()  // Gets info about the loaded game
	s.NWAAutoSplitter.EmuStatus()    // Gets the status of the emu
	s.NWAAutoSplitter.ClientID()     // Provides the client name to the NWA interface
	s.NWAAutoSplitter.CoreInfo()     // Might be useful to display the platform & core names
	s.NWAAutoSplitter.CoreMemories() // Get info about the memory banks available

	// this is the core loop of autosplitting
	// queries the device (emu, hardware, application) at the rate specified in ms
	for {
		processStart := time.Now()

		fmt.Printf("Checking for autosplitting updates.\n")
		autoState, err2 := s.NWAAutoSplitter.Update()
		if err2 != nil {
			return
		}
		if autoState.Reset {
			//restart run
			commandDispatcher.Dispatch(dispatcher.RESET, nil)
		}
		if autoState.Split {
			//split run
			commandDispatcher.Dispatch(dispatcher.SPLIT, nil)
		}
		// TODO: Close the connection after closing the splits file or receiving a disconnect signal
		// s.NWAAutoSplitter.Client.Close()
		processElapsed := time.Since(processStart)
		// fmt.Println(processStart)
		// fmt.Println(processElapsed)
		time.Sleep(min(mil, max(0, mil-processElapsed)))
	}
}

func (s Splitters) processQUSB2SNES(commandDispatcher *dispatcher.Service) {
	// 	// //QUSB2SNES example
	// if QUSB2SNESAutoSplitterService != nil {
	// 	// client.SetName("annelid")

	// 	// version, err := client.AppVersion()
	// 	// fmt.Printf("Server version is %v\n", version)

	// 	// devices, err := client.ListDevice()

	// 	// if len(devices) != 1 {
	// 	// 	if len(devices) == 0 {
	// 	// 		return errors.New("no devices present")
	// 	// 	}
	// 	// 	return errors.Errorf("unexpected devices: %#v", devices)
	// 	// }
	// 	// device := devices[0]
	// 	// fmt.Printf("Using device %v\n", device)

	// 	// client.Attach(device)
	// 	// fmt.Println("Connected.")

	// 	// info, err := client.Info()
	// 	// fmt.Printf("%#v\n", info)

	// 	// var autosplitter AutoSplitter = NewSuperMetroidAutoSplitter(settings)

	// 	// for {
	// 	// 	summary, err := autosplitter.Update(client)
	// 	// 	if summary.Start {
	// 	// 		timer.Start()
	// 	// 	}
	// 	// 	if summary.Reset {
	// 	// 		if resetTimerOnGameReset == true {
	// 	// 			timer.Reset(true)
	// 	// 		}
	// 	// 	}
	// 	// 	if summary.Split {
	// 	// 		// IGT
	// 	// 		timer.SetGameTime(*t)
	// 	// 		// RTA
	// 	// 		timer.Split()
	// 	// 	}

	// 	// 	if ev == TimerReset {
	// 	// 		// creates a new SNES state
	// 	// 		autosplitter.ResetGameTracking()
	// 	// 		if resetGameOnTimerReset == true {
	// 	// 			client.Reset()
	// 	// 		}
	// 	// 	}

	// 	// 	time.Sleep(time.Duration(float64(time.Second) / pollingRate))
	// 	// }
	// }
}
