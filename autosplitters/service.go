package autosplitters

import (
	nwa "github.com/zellydev-games/opensplit/autosplitters/NWA"
)

type AutosplitterType int

const (
	NWA AutosplitterType = iota
	QUSB2SNES
)

// type Splitter interface {}

// need a return type that can handle any type we give it
func NewService(UseAutosplitter bool, ResetTimerOnGameReset bool, Port uint32, Addr string /*game Game,*/, Type AutosplitterType) *nwa.NWASplitter {
	if UseAutosplitter {
		if Type == NWA {
			client, _ := nwa.Connect(Addr, Port)
			return &nwa.NWASplitter{
				ResetTimerOnGameReset: ResetTimerOnGameReset,
				Client:                *client,
			}
		}
		// if Type == QUSB2SNES {
		// 	client, err := qusb2snes.Connect()
		// 	if err != nil {
		// 		return err
		// 	}
		// 	return &client
		// }
	}
	return nil
}
