package nwa

// TODO: handle errors correctly

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// public
type NWASplitter struct {
	ResetTimerOnGameReset bool
	Client                NWASyncClient
	nwaMemory             []MemoryEntry
	resetConditions       [][]Element
	splitConditions       [][]Element
}

// type compare_type int

// const (
// 	ceqp compare_type = iota // current value equal to prior value
// 	ceqe                     // current value equal to expected value
// 	cnep                     // current value not equal to prior value
// 	cnee                     // current value not equal to expected value
// 	cgtp                     // current value greater than prior value
// 	cgte                     // current value greater than expected value
// 	cltp                     // current value less than than prior value
// 	clte                     // current value less than than expected value
// 	eeqc                     // expected value equal to current value
// 	eeqp                     // expected value equal to prior value
// 	enec                     // expected value not equal to current value
// 	enep                     // expected value not equal to prior value
// 	egtc                     // expected value greater than current value
// 	egtp                     // expected value greater than prior value
// 	eltc                     // expected value less than than current value
// 	eltp                     // expected value less than than prior value
// 	peqc                     // prior value equal to current value
// 	peqe                     // prior value equal to expected value
// 	pnec                     // prior value not equal to current value
// 	pnee                     // prior value not equal to expected value
// 	pgtc                     // prior value greater than current value
// 	pgte                     // prior value greater than expected value
// 	pltc                     // prior value less than than current value
// 	plte                     // prior value less than than expected value
// )

type Element struct {
	memoryEntryName string
	expectedValue   *int
	compareType     string
}

type MemoryEntry struct {
	Name         string
	MemoryBank   string
	Address      string
	Size         string
	currentValue *int
	priorValue   *int
}

// Setup the memory map being read by the NWA splitter and the maps for the reset and split conditions
func (b *NWASplitter) MemAndConditionsSetup(memData []string, resetConditionImport []string, splitConditionImport []string) {
	// Populate Start Condition List
	for _, p := range memData {
		// create memory entry
		memName := strings.Split(p, ",")

		entry := MemoryEntry{
			Name:       memName[0],
			MemoryBank: memName[1],
			Address:    memName[2],
			Size:       memName[3]}
		// add memory map entries to nwaMemory list
		b.nwaMemory = append(b.nwaMemory, entry)
	}

	// Populate Reset Condition List
	for _, p := range resetConditionImport {
		var condition []Element
		// create elements
		// add elements to condition list
		resetCon := strings.Split(strings.Split(p, ":")[1], " ")
		for _, q := range resetCon {
			elements := strings.Split(q, ",")

			if len(elements) == 3 {
				cT := elements[2]

				// convert hex string to int
				num, err := strconv.ParseUint(elements[1], 16, 64)
				integer := int(num)
				if err != nil {
					log.Fatalf("Failed to convert string to integer: %v", err)
				}
				intPtr := new(int)
				*intPtr = integer

				condition = append(condition, Element{
					memoryEntryName: elements[0],
					expectedValue:   intPtr,
					compareType:     cT})
			} else if len(elements) == 2 {
				cT := elements[1]

				condition = append(condition, Element{
					memoryEntryName: elements[0],
					compareType:     cT})
			} else {
				fmt.Printf("Too many or too few conditions given: %#v\n", q)
			}
		}
		// add condition lists to Reset Conditions list
		b.resetConditions = append(b.resetConditions, condition)
	}

	// Populate Split Condition List
	for _, p := range splitConditionImport {
		var condition []Element
		// create elements
		// add elements to condition list
		splitCon := strings.Split(strings.Split(p, ":")[1], " ")
		for _, q := range splitCon {
			elements := strings.Split(q, ",")

			if len(elements) == 3 {
				cT := elements[2]

				num, err := strconv.ParseUint(elements[1], 16, 64)
				integer := int(num)
				if err != nil {
					log.Fatalf("Failed to convert string to integer: %v", err)
				}
				intPtr := new(int)
				*intPtr = integer

				condition = append(condition, Element{
					memoryEntryName: elements[0],
					expectedValue:   intPtr,
					compareType:     cT})
			} else if len(elements) == 2 {
				cT := elements[1]

				condition = append(condition, Element{
					memoryEntryName: elements[0],
					compareType:     cT})
			} else {
				fmt.Printf("Too many or too few conditions given: %#v\n", q)
			}
		}
		// add condition lists to Split Conditions list
		b.splitConditions = append(b.splitConditions, condition)
	}
}

func (b *NWASplitter) ClientID() {
	cmd := "MY_NAME_IS"
	args := "OpenSplit"
	summary, err := b.Client.ExecuteCommand(cmd, &args)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) EmuInfo() {
	cmd := "EMULATOR_INFO"
	args := "0"
	summary, err := b.Client.ExecuteCommand(cmd, &args)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) EmuGameInfo() {
	cmd := "GAME_INFO"
	summary, err := b.Client.ExecuteCommand(cmd, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) EmuStatus() {
	cmd := "EMULATION_STATUS"
	summary, err := b.Client.ExecuteCommand(cmd, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) CoreInfo() {
	cmd := "CORE_CURRENT_INFO"
	summary, err := b.Client.ExecuteCommand(cmd, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) CoreMemories() {
	cmd := "CORE_MEMORIES"
	summary, err := b.Client.ExecuteCommand(cmd, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", summary)
}

// currently only supports 1 byte reads
func (b *NWASplitter) Update() (nwaSummary, error) {
	cmd := "CORE_READ"
	for i, p := range b.nwaMemory {
		args := p.MemoryBank + ";" + p.Address + ";" + p.Size
		summary, err := b.Client.ExecuteCommand(cmd, &args)
		if err != nil {
			return nwaSummary{}, err
		}
		fmt.Printf("%#v\n", summary)

		b.nwaMemory[i].priorValue = b.nwaMemory[i].currentValue
		switch v := summary.(type) {
		case []byte:

			// need to handle more than 1 byte at a time
			// if len(v) == 1 {
			// 	val := int(v[0])
			// 	b.nwaMemory[i].currentValue = &val
			// } else if len(v) > 1 {
			// length 1
			var temp_int uint8
			//length 2
			// var temp_int uint16
			// length 4
			// var temp_int uint32
			// length 8
			// var temp_int uint64
			err := binary.Read(bytes.NewReader(v), binary.LittleEndian, &temp_int)
			if err != nil {
				fmt.Println("Error reading binary data:", err)
			}
			integer := int(temp_int)
			b.nwaMemory[i].currentValue = &integer
			// }
		case NWAError:
			fmt.Printf("%#v\n", v)
		default:
			fmt.Printf("%#v\n", v)
		}
	}

	reset := b.reset()
	split := b.split()

	return nwaSummary{
		Reset: reset,
		Split: split,
	}, nil
}

// private
type nwaSummary struct {
	Reset bool
	Split bool
}

// Checks conditions and returns reset state
func (b *NWASplitter) reset() bool {
	fmt.Printf("Checking reset state\n")
	if b.ResetTimerOnGameReset {
		for _, p := range b.resetConditions {
			resetState := true
			var tempstate bool
			for _, q := range p {
				index, found := b.findInSlice(b.nwaMemory, q.memoryEntryName)
				if found {
					tempstate = compare(q.compareType, b.nwaMemory[index].currentValue, b.nwaMemory[index].priorValue, q.expectedValue)
				} else {
					fmt.Printf("How did you get here?\n")
				}
				resetState = resetState && tempstate
			}
			if resetState {
				fmt.Printf("Time to reset\n")
				return true
			}
		}
		return false
	} else {
		return false
	}
}

// Checks conditions and returns split state
func (b *NWASplitter) split() bool {
	for _, p := range b.splitConditions {
		fmt.Printf("Checking split state\n")
		splitState := true
		var tempstate bool
		for _, q := range p {
			index, found := b.findInSlice(b.nwaMemory, q.memoryEntryName)
			if found {
				tempstate = compare(q.compareType, b.nwaMemory[index].currentValue, b.nwaMemory[index].priorValue, q.expectedValue)
			} else {
				fmt.Printf("How did you get here?\n")
			}
			splitState = splitState && tempstate
		}
		if splitState {
			fmt.Printf("Time to split\n")
			return true
		}
	}
	return false
}

func (b *NWASplitter) findInSlice(slice []MemoryEntry, target string) (int, bool) {
	for i, v := range slice {
		if v.Name == target {
			return i, true // Return index and true if found
		}
	}
	return -1, false // Return -1 and false if not found
}

func compare(input string, current *int, prior *int, expected *int) bool {
	switch input {
	case "ceqp":
		fallthrough
	case "peqc":
		if (prior == nil) || (current == nil) {
			return false
		} else {
			return *prior == *current
		}
	case "ceqe":
		fallthrough
	case "eeqc":
		if (expected == nil) || (current == nil) {
			return false
		} else {
			return *expected == *current
		}
	case "eeqp":
		fallthrough
	case "peqe":
		if (expected == nil) || (prior == nil) {
			return false
		} else {
			return *prior == *expected
		}
	case "cnep":
		fallthrough
	case "pnec":
		if (prior == nil) || (current == nil) {
			return false
		} else {
			return *prior != *current
		}
	case "cnee":
		fallthrough
	case "enec":
		if (expected == nil) || (current == nil) {
			return false
		} else {
			return *expected != *current
		}
	case "enep":
		fallthrough
	case "pnee":
		if (expected == nil) || (prior == nil) {
			return false
		} else {
			return *prior != *expected
		}
	case "cgtp":
		fallthrough
	case "pltc":
		if (prior == nil) || (current == nil) {
			return false
		} else {
			return *prior < *current
		}
	case "cgte":
		fallthrough
	case "eltc":
		if (expected == nil) || (current == nil) {
			return false
		} else {
			return *expected < *current
		}
	case "egtp":
		fallthrough
	case "plte":
		if (expected == nil) || (prior == nil) {
			return false
		} else {
			return *prior < *expected
		}
	case "cltp":
		fallthrough
	case "pgtc":
		if (prior == nil) || (current == nil) {
			return false
		} else {
			return *prior > *current
		}
	case "clte":
		fallthrough
	case "egtc":
		if (expected == nil) || (current == nil) {
			return false
		} else {
			return *expected > *current
		}
	case "eltp":
		fallthrough
	case "pgte":
		if (expected == nil) || (prior == nil) {
			return false
		} else {
			return *prior > *expected
		}
	default:
		return false
	}
}
