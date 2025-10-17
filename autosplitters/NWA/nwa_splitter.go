package nwa

import (
	"fmt"
)

type NWASummary struct {
	Start bool
	Reset bool
	Split bool
}

type NWASplitter struct {
	priorLevel            uint8
	level                 uint8
	resetTimerOnGameReset bool
	client                NWASyncClient
}

func (b *NWASplitter) ClientID() {
	cmd := "MY_NAME_IS"
	args := "Annelid"
	summary, err := b.client.ExecuteCommand(cmd, &args)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) EmuInfo() {
	cmd := "EMULATOR_INFO"
	args := "0"
	summary, err := b.client.ExecuteCommand(cmd, &args)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) EmuGameInfo() {
	cmd := "GAME_INFO"
	summary, err := b.client.ExecuteCommand(cmd, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) EmuStatus() {
	cmd := "EMULATION_STATUS"
	summary, err := b.client.ExecuteCommand(cmd, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) CoreInfo() {
	cmd := "CORE_CURRENT_INFO"
	summary, err := b.client.ExecuteCommand(cmd, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) CoreMemories() {
	cmd := "CORE_MEMORIES"
	summary, err := b.client.ExecuteCommand(cmd, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) Update() (NWASummary, error) {
	b.priorLevel = b.level
	cmd := "CORE_READ"
	args := "RAM;$0010;1"
	summary, err := b.client.ExecuteCommand(cmd, &args)
	if err != nil {
		return NWASummary{}, err
	}
	fmt.Printf("%#v\n", summary)

	switch v := summary.(type) {
	case []byte:
		if len(v) > 0 {
			b.level = v[0]
		}
	case NWAError:
		fmt.Printf("%#v\n", v)
	default:
		fmt.Printf("%#v\n", v)
	}

	fmt.Printf("%#v\n", b.level)

	start := b.Start()
	reset := b.Reset()
	split := b.Split()

	return NWASummary{
		Start: start,
		Reset: reset,
		Split: split,
	}, nil
}

func (b *NWASplitter) Start() bool {
	return b.level == 1 && b.priorLevel == 0
}

func (b *NWASplitter) Reset() bool {
	return b.level == 0 &&
		b.priorLevel != 0 &&
		b.resetTimerOnGameReset
}

func (b *NWASplitter) Split() bool {
	return b.level > b.priorLevel && b.priorLevel < 100
}
