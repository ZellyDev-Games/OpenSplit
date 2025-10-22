package nwa

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
	startConditions       [][]Element
	resetConditions       [][]Element
	splitConditions       [][]Element
}

type compare_type int

const (
	ceqp compare_type = iota // current value equal to prior value
	ceqe                     // current value equal to expected value
	cnep                     // current value not equal to prior value
	cnee                     // current value not equal to expected value
	cgtp                     // current value greater than prior value
	cgte                     // current value greater than expected value
	cltp                     // current value less than than prior value
	clte                     // current value less than than expected value
	eeqc                     // expected value equal to current value
	eeqp                     // expected value equal to prior value
	enec                     // expected value not equal to current value
	enep                     // expected value not equal to prior value
	egtc                     // expected value greater than current value
	egtp                     // expected value greater than prior value
	eltc                     // expected value less than than current value
	eltp                     // expected value less than than prior value
	peqc                     // prior value equal to current value
	peqe                     // prior value equal to expected value
	pnec                     // prior value not equal to current value
	pnee                     // prior value not equal to expected value
	pgtc                     // prior value greater than current value
	pgte                     // prior value greater than expected value
	pltc                     // prior value less than than current value
	plte                     // prior value less than than expected value
	cter                     // compare type error
)

type Element struct {
	// name            string
	memoryEntryName string
	expectedValue   *int
	compareType     compare_type
}

type MemoryEntry struct {
	name         string
	memoryBank   string
	memory       string
	size         string
	currentValue *int
	priorValue   *int
}

func (b *NWASplitter) MemAndConditionsSetup(memData []string, startConditionImport []string, resetConditionImport []string, splitConditionImport []string) {
	// Populate Start Condition List
	for _, p := range memData {
		// create memory entry
		memName := strings.Split(p, ",")
		// integer, err := strconv.Atoi(memName[2]) // Atoi returns an int and an error

		// if err != nil {
		// log.Fatalf("Failed to convert string to integer: %v", err)
		// }

		entry := MemoryEntry{
			name:       memName[0],
			memoryBank: memName[1],
			memory:     memName[2],
			size:       memName[3]}
		// add memory map entries to nwaMemory list
		b.nwaMemory = append(b.nwaMemory, entry)
	}

	// Populate Start Condition List
	for _, p := range startConditionImport {
		var condition []Element
		// create elements
		// add elements to condition list
		startCon := strings.Split(p, ",")
		if len(startCon) != 2 || len(startCon) != 3 {
			// Error. Too many or too few elements
		} else {
			// convert string compare type to enum
			cT := compareTypeConverter(startCon[2])
			if cT == cter {
				// return an error
			}

			if len(startCon) == 3 {
				integer, err := strconv.Atoi(startCon[1]) // Atoi returns an int and an error
				if err != nil {
					log.Fatalf("Failed to convert string to integer: %v", err)
				}
				intPtr := new(int)
				*intPtr = integer

				condition = append(condition, Element{
					memoryEntryName: startCon[0],
					expectedValue:   intPtr,
					compareType:     cT})
			} else if len(startCon) == 2 {
				condition = append(condition, Element{
					memoryEntryName: startCon[0],
					compareType:     cT})
			}
			// add condition lists to StartConditions list
			b.startConditions = append(b.startConditions, condition)
		}
	}
	// Populate Reset Condition List
	for _, p := range resetConditionImport {
		var condition []Element
		// create elements
		// add elements to condition list
		resetCon := strings.Split(p, ",")
		if len(resetCon) != 2 || len(resetCon) != 3 {
			// Error. Too many or too few elements
		} else {
			// convert string compare type to enum
			cT := compareTypeConverter(resetCon[2])
			if cT == cter {
				// return an error
			}

			if len(resetCon) == 3 {
				integer, err := strconv.Atoi(resetCon[1]) // Atoi returns an int and an error
				if err != nil {
					log.Fatalf("Failed to convert string to integer: %v", err)
				}
				intPtr := new(int)
				*intPtr = integer

				condition = append(condition, Element{
					memoryEntryName: resetCon[0],
					expectedValue:   intPtr,
					compareType:     cT})
			} else if len(resetCon) == 2 {
				condition = append(condition, Element{
					memoryEntryName: resetCon[0],
					compareType:     cT})
			}
			// add condition lists to StartConditions list
			b.resetConditions = append(b.resetConditions, condition)
		}
	}

	// Populate Split Condition List
	for _, p := range splitConditionImport {
		var condition []Element
		// create elements
		// add elements to condition list
		splitCon := strings.Split(p, ",")
		if len(splitCon) != 2 || len(splitCon) != 3 {
			// Error. Too many or too few elements
		} else {
			// convert string compare type to enum
			cT := compareTypeConverter(splitCon[2])
			if cT == cter {
				// return an error
			}

			if len(splitCon) == 3 {
				integer, err := strconv.Atoi(splitCon[1]) // Atoi returns an int and an error
				if err != nil {
					log.Fatalf("Failed to convert string to integer: %v", err)
				}
				intPtr := new(int)
				*intPtr = integer

				condition = append(condition, Element{
					memoryEntryName: splitCon[0],
					expectedValue:   intPtr,
					compareType:     cT})
			} else if len(splitCon) == 2 {
				condition = append(condition, Element{
					memoryEntryName: splitCon[0],
					compareType:     cT})
			}
			// add condition lists to StartConditions list
			b.splitConditions = append(b.splitConditions, condition)
		}
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

func (b *NWASplitter) Update() (nwaSummary, error) {
	cmd := "CORE_READ"
	for _, p := range b.nwaMemory {
		args := p.memoryBank + ";" + p.memory + ";" + p.size
		summary, err := b.Client.ExecuteCommand(cmd, &args)
		if err != nil {
			return nwaSummary{}, err
		}
		fmt.Printf("%#v\n", summary)

		switch v := summary.(type) {
		case []byte:
			if len(v) == 1 {
				*p.currentValue = int(v[0])
			} else if len(v) > 1 {
				var i int
				buf := bytes.NewReader(v)
				err := binary.Read(buf, binary.LittleEndian, &i)
				if err != nil {
					fmt.Println("Error reading binary data:", err)
				}
				*p.currentValue = i
			}
		case NWAError:
			fmt.Printf("%#v\n", v)
		default:
			fmt.Printf("%#v\n", v)
		}
	}

	start := b.start()
	reset := b.reset()
	split := b.split()

	return nwaSummary{
		Start: start,
		Reset: reset,
		Split: split,
	}, nil
}

// private
type nwaSummary struct {
	Start bool
	Reset bool
	Split bool
}

func (b *NWASplitter) start() bool {
	startState := true
	for _, p := range b.startConditions {
		var tempstate bool
		for _, q := range p {
			index, found := b.findInSlice(b.nwaMemory, q.memoryEntryName)
			if found {
				tempstate = compare(q.compareType, b.nwaMemory[index].currentValue, b.nwaMemory[index].priorValue, q.expectedValue)
			} else {
				// throw error
			}
			startState = startState && tempstate
		}
		if startState {
			return true
		}
	}
	return false
}

func (b *NWASplitter) reset() bool {
	if b.ResetTimerOnGameReset {
		resetState := true
		for _, p := range b.resetConditions {
			var tempstate bool
			for _, q := range p {
				index, found := b.findInSlice(b.nwaMemory, q.memoryEntryName)
				if found {
					tempstate = compare(q.compareType, b.nwaMemory[index].currentValue, b.nwaMemory[index].priorValue, q.expectedValue)
				} else {
					// throw error
				}
				resetState = resetState && tempstate
			}
			if resetState {
				return true
			}
		}
		return false
	} else {
		return false
	}
}

func (b *NWASplitter) split() bool {
	splitState := true
	for _, p := range b.splitConditions {
		var tempstate bool
		for _, q := range p {
			index, found := b.findInSlice(b.nwaMemory, q.memoryEntryName)
			if found {
				tempstate = compare(q.compareType, b.nwaMemory[index].currentValue, b.nwaMemory[index].priorValue, q.expectedValue)
			} else {
				// throw error
			}
			splitState = splitState && tempstate
		}
		if splitState {
			return true
		}
	}
	return false
}

func (b *NWASplitter) findInSlice(slice []MemoryEntry, target string) (int, bool) {
	for i, v := range slice {
		if v.name == target {
			return i, true // Return index and true if found
		}
	}
	return -1, false // Return -1 and false if not found
}

func compareTypeConverter(input string) compare_type {
	switch input {
	case "ceqp":
		return ceqp
	case "ceqe":
		return ceqe
	case "cnep":
		return cnep
	case "cnee":
		return cnee
	case "cgtp":
		return cgtp
	case "cgte":
		return cgte
	case "cltp":
		return cltp
	case "clte":
		return clte
	case "eeqc":
		return eeqc
	case "eeqp":
		return eeqp
	case "enec":
		return enec
	case "enep":
		return enep
	case "egtc":
		return egtc
	case "egtp":
		return egtp
	case "eltc":
		return eltc
	case "eltp":
		return eltp
	case "peqc":
		return peqc
	case "peqe":
		return peqe
	case "pnec":
		return pnec
	case "pnee":
		return pnee
	case "pgtc":
		return pgtc
	case "pgte":
		return pgte
	case "pltc":
		return pltc
	case "plte":
		return plte
	default:
		return cter
	}
}

func compare(input compare_type, current *int, prior *int, expected *int) bool {
	switch input {
	case ceqp:
		return *current == *prior
	case ceqe:
		return *current == *expected
	case cnep:
		return *current != *prior
	case cnee:
		return *current != *expected
	case cgtp:
		return *current > *prior
	case cgte:
		return *current > *expected
	case cltp:
		return *current < *prior
	case clte:
		return *current < *expected
	case eeqc:
		return *expected == *current
	case eeqp:
		return *expected == *prior
	case enec:
		return *expected != *current
	case enep:
		return *expected != *prior
	case egtc:
		return *expected > *current
	case egtp:
		return *expected > *prior
	case eltc:
		return *expected < *current
	case eltp:
		return *expected < *prior
	case peqc:
		return *prior == *current
	case peqe:
		return *prior == *expected
	case pnec:
		return *prior != *current
	case pnee:
		return *prior != *expected
	case pgtc:
		return *prior > *current
	case pgte:
		return *prior > *expected
	case pltc:
		return *prior < *current
	case plte:
		return *prior < *expected
	default:
		return false
	}
}
