package nwa

import (
	"encoding/binary"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// public
type NWASplitter struct {
	Client          NWASyncClient
	nwaMemory       []memoryWatcher
	startConditions []conditionList
	resetConditions []conditionList
	splitConditions []conditionList
}

type element struct {
	memoryEntryName string
	expectedValue   *int
	compareType     string
	result          compareFunc
}

type compareFunc func(input string, prior *int, current *int, expected *int) bool

type memoryWatcher struct {
	name         string
	memoryBank   string
	address      string
	size         string
	currentValue *int
	priorValue   *int
}

type conditionList struct {
	Name   string
	memory []element
}

// Setup the memory map being read by the NWA splitter and the maps for the reset and split conditions
func (b *NWASplitter) MemAndConditionsSetup(memData []string, startConditionImport []string, resetConditionImport []string, splitConditionImport []string) {
	delimiter1 := "="
	delimiter2 := "≠"
	delimiter3 := "<"
	delimiter4 := ">"
	delimiter5 := "&"
	delimiter6 := "|"
	delimiter7 := "^"
	compareStringCurrent := "current"
	compareStringPrior := "prior"

	for _, p := range memData {
		mem := strings.Split(p, ",")

		entry := memoryWatcher{
			name:         mem[0],
			memoryBank:   mem[1],
			address:      mem[2],
			size:         mem[3],
			currentValue: new(int),
			priorValue:   new(int),
		}
		b.nwaMemory = append(b.nwaMemory, entry)
	}

	// Populate Start Condition List
	for _, p := range startConditionImport {
		var condition conditionList
		// create elements
		// add elements to reset condition list
		startName := strings.Split(p, ":")[0]
		startCon := strings.Split(strings.Split(p, ":")[1], " ")

		condition.Name = startName

		for _, q := range startCon {
			if strings.Contains(q, "&&") {
				continue
			}

			var tempElement element

			components := strings.Split(q, ",")

			tempElement.memoryEntryName = components[0]
			tempElement.result = compare

			if strings.Contains(components[1], "=") {
				compStrings := strings.Split(components[1], delimiter1)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "ceqe"
					case compareStringPrior:
						tempElement.compareType = "peqe"
					}
				} else {
					if compStrings[0] == compareStringCurrent && compStrings[1] == compareStringPrior {
						tempElement.compareType = "ceqp"
					} else {
						tempElement.compareType = "peqc"
					}
				}
			} else if strings.Contains(components[1], "≠") {
				compStrings := strings.Split(components[1], delimiter2)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "cnee"
					case compareStringPrior:
						tempElement.compareType = "pnee"
					}
				} else {
					if compStrings[0] == compareStringCurrent && compStrings[1] == compareStringPrior {
						tempElement.compareType = "cnep"
					} else {
						tempElement.compareType = "pnec"
					}
				}
			} else if strings.Contains(components[1], "<") {
				compStrings := strings.Split(components[1], delimiter3)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "clte"
					case compareStringPrior:
						tempElement.compareType = "plte"
					}
				} else {
					if compStrings[0] == compareStringCurrent && compStrings[1] == compareStringPrior {
						tempElement.compareType = "cltp"
					} else {
						tempElement.compareType = "pltc"
					}
				}
			} else if strings.Contains(components[1], ">") {
				compStrings := strings.Split(components[1], delimiter4)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "cgte"
					case compareStringPrior:
						tempElement.compareType = "pgte"
					}
				} else {
					if compStrings[0] == compareStringCurrent && compStrings[1] == compareStringPrior {
						tempElement.compareType = "cgtp"
					} else {
						tempElement.compareType = "pgtc"
					}
				}
			} else if strings.Contains(components[1], "&") {
				compStrings := strings.Split(components[1], delimiter5)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "cbae"
					case compareStringPrior:
						tempElement.compareType = "pbae"
					}
				}
			} else if strings.Contains(components[1], "|") {
				compStrings := strings.Split(components[1], delimiter6)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "cboe"
					case compareStringPrior:
						tempElement.compareType = "pboe"
					}
				}
			} else if strings.Contains(components[1], "^") {
				compStrings := strings.Split(components[1], delimiter7)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "cbne"
					case compareStringPrior:
						tempElement.compareType = "pbne"
					}
				}
			}

			condition.memory = append(condition.memory, tempElement)
		}
		// add condition lists to Start Conditions list
		b.startConditions = append(b.startConditions, condition)
	}

	// Populate Reset Condition List
	for _, p := range resetConditionImport {
		var condition conditionList
		// create elements
		// add elements to reset condition list
		resetName := strings.Split(p, ":")[0]
		resetCon := strings.Split(strings.Split(p, ":")[1], " ")

		condition.Name = resetName

		for _, q := range resetCon {
			if strings.Contains(q, "&&") {
				continue
			}

			var tempElement element

			components := strings.Split(q, ",")

			tempElement.memoryEntryName = components[0]
			tempElement.result = compare

			if strings.Contains(components[1], "=") {
				compStrings := strings.Split(components[1], delimiter1)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "ceqe"
					case compareStringPrior:
						tempElement.compareType = "peqe"
					}
				} else {
					if compStrings[0] == compareStringCurrent && compStrings[1] == compareStringPrior {
						tempElement.compareType = "ceqp"
					} else {
						tempElement.compareType = "peqc"
					}
				}
			} else if strings.Contains(components[1], "≠") {
				compStrings := strings.Split(components[1], delimiter2)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "cnee"
					case compareStringPrior:
						tempElement.compareType = "pnee"
					}
				} else {
					if compStrings[0] == compareStringCurrent && compStrings[1] == compareStringPrior {
						tempElement.compareType = "cnep"
					} else {
						tempElement.compareType = "pnec"
					}
				}
			} else if strings.Contains(components[1], "<") {
				compStrings := strings.Split(components[1], delimiter3)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "clte"
					case compareStringPrior:
						tempElement.compareType = "plte"
					}
				} else {
					if compStrings[0] == compareStringCurrent && compStrings[1] == compareStringPrior {
						tempElement.compareType = "cltp"
					} else {
						tempElement.compareType = "pltc"
					}
				}
			} else if strings.Contains(components[1], ">") {
				compStrings := strings.Split(components[1], delimiter4)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "cgte"
					case compareStringPrior:
						tempElement.compareType = "pgte"
					}
				} else {
					if compStrings[0] == compareStringCurrent && compStrings[1] == compareStringPrior {
						tempElement.compareType = "cgtp"
					} else {
						tempElement.compareType = "pgtc"
					}
				}
			} else if strings.Contains(components[1], "&") {
				compStrings := strings.Split(components[1], delimiter5)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "cbae"
					case compareStringPrior:
						tempElement.compareType = "pbae"
					}
				}
			} else if strings.Contains(components[1], "|") {
				compStrings := strings.Split(components[1], delimiter6)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "cboe"
					case compareStringPrior:
						tempElement.compareType = "pboe"
					}
				}
			} else if strings.Contains(components[1], "^") {
				compStrings := strings.Split(components[1], delimiter7)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "cbne"
					case compareStringPrior:
						tempElement.compareType = "pbne"
					}
				}
			}

			condition.memory = append(condition.memory, tempElement)
		}
		// add condition lists to Reset Conditions list
		b.resetConditions = append(b.resetConditions, condition)
	}

	// Populate Split Condition List
	for _, p := range splitConditionImport {
		var condition conditionList
		// create elements
		// add elements to split condition list
		splitName := strings.Split(p, ":")[0]
		splitCon := strings.Split(strings.Split(p, ":")[1], " ")

		condition.Name = splitName

		for _, q := range splitCon {
			if strings.Contains(q, "&&") {
				continue
			}

			var tempElement element

			components := strings.Split(q, ",")

			tempElement.memoryEntryName = components[0]
			tempElement.result = compare

			if strings.Contains(components[1], "=") {
				compStrings := strings.Split(components[1], delimiter1)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "ceqe"
					case compareStringPrior:
						tempElement.compareType = "peqe"
					}
				} else {
					if compStrings[0] == compareStringCurrent && compStrings[1] == compareStringPrior {
						tempElement.compareType = "ceqp"
					} else {
						tempElement.compareType = "peqc"
					}
				}
			} else if strings.Contains(components[1], "≠") {
				compStrings := strings.Split(components[1], delimiter2)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "cnee"
					case compareStringPrior:
						tempElement.compareType = "pnee"
					}
				} else {
					if compStrings[0] == compareStringCurrent && compStrings[1] == compareStringPrior {
						tempElement.compareType = "cnep"
					} else {
						tempElement.compareType = "pnec"
					}
				}
			} else if strings.Contains(components[1], "<") {
				compStrings := strings.Split(components[1], delimiter3)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "clte"
					case compareStringPrior:
						tempElement.compareType = "plte"
					}
				} else {
					if compStrings[0] == compareStringCurrent && compStrings[1] == compareStringPrior {
						tempElement.compareType = "cltp"
					} else {
						tempElement.compareType = "pltc"
					}
				}
			} else if strings.Contains(components[1], ">") {
				compStrings := strings.Split(components[1], delimiter4)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "cgte"
					case compareStringPrior:
						tempElement.compareType = "pgte"
					}
				} else {
					if compStrings[0] == compareStringCurrent && compStrings[1] == compareStringPrior {
						tempElement.compareType = "cgtp"
					} else {
						tempElement.compareType = "pgtc"
					}
				}
			} else if strings.Contains(components[1], "&") {
				compStrings := strings.Split(components[1], delimiter5)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "cbae"
					case compareStringPrior:
						tempElement.compareType = "pbae"
					}
				}
			} else if strings.Contains(components[1], "|") {
				compStrings := strings.Split(components[1], delimiter6)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "cboe"
					case compareStringPrior:
						tempElement.compareType = "pboe"
					}
				}
			} else if strings.Contains(components[1], "^") {
				compStrings := strings.Split(components[1], delimiter7)
				tempElement.expectedValue = hexToInt(compStrings[1])
				if tempElement.expectedValue != nil {
					switch compStrings[0] {
					case compareStringCurrent:
						tempElement.compareType = "cbne"
					case compareStringPrior:
						tempElement.compareType = "pbne"
					}
				}
			}

			condition.memory = append(condition.memory, tempElement)
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
		// panic(err)
		println(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) EmuInfo() {
	cmd := "EMULATOR_INFO"
	args := "0"
	summary, err := b.Client.ExecuteCommand(cmd, &args)
	if err != nil {
		println(err)
		// panic(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) EmuGameInfo() {
	cmd := "GAME_INFO"
	summary, err := b.Client.ExecuteCommand(cmd, nil)
	if err != nil {
		// panic(err)
		println(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) EmuStatus() {
	cmd := "EMULATION_STATUS"
	summary, err := b.Client.ExecuteCommand(cmd, nil)
	if err != nil {
		// panic(err)
		println(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) CoreInfo() {
	cmd := "CORE_CURRENT_INFO"
	summary, err := b.Client.ExecuteCommand(cmd, nil)
	if err != nil {
		// panic(err)
		println(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) CoreMemories() {
	cmd := "CORE_MEMORIES"
	summary, err := b.Client.ExecuteCommand(cmd, nil)
	if err != nil {
		// panic(err)
		println(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) SoftResetConsole() {
	cmd := "EMULATION_RESET"
	summary, err := b.Client.ExecuteCommand(cmd, nil)
	if err != nil {
		// panic(err)
		println(err)
	}
	fmt.Printf("%#v\n", summary)
}

func (b *NWASplitter) HardResetConsole() {
	// cmd := "EMULATION_STOP"
	cmd := "EMULATION_RELOAD"
	summary, err := b.Client.ExecuteCommand(cmd, nil)
	if err != nil {
		// panic(err)
		println(err)
	}
	fmt.Printf("%#v\n", summary)
}

// currently only suppports 1 memory source at a time
// likely WRAM for SNES and RAM for NES
func (b *NWASplitter) Update(splitIndex int) (nwaSummary, error) {

	cmd := "CORE_READ"
	domain := b.nwaMemory[0].memoryBank
	var requestString string

	for _, watcher := range b.nwaMemory {
		requestString += ";" + watcher.address + ";" + watcher.size
		*watcher.priorValue = *watcher.currentValue
	}

	args := domain + requestString
	summary, err := b.Client.ExecuteCommand(cmd, &args)
	if err != nil {
		return nwaSummary{}, err
	}
	fmt.Printf("%#v\n", summary)

	switch v := summary.(type) {
	case []byte:
		// update memoryWatcher with data
		runningTotal := 0
		for _, watcher := range b.nwaMemory {
			size, _ := strconv.Atoi(watcher.size)
			switch size {
			case 1:
				*watcher.currentValue = int(v[runningTotal])
				runningTotal += size
			case 2:
				*watcher.currentValue = int(binary.LittleEndian.Uint16(v[runningTotal : runningTotal+size]))
				runningTotal += size
			case 3:
				fallthrough
			case 4:
				*watcher.currentValue = int(binary.LittleEndian.Uint32(v[runningTotal : runningTotal+size]))
				runningTotal += size
			case 5:
				fallthrough
			case 6:
				fallthrough
			case 7:
				fallthrough
			case 8:
				*watcher.currentValue = int(binary.LittleEndian.Uint64(v[runningTotal : runningTotal+size]))
				runningTotal += size
			}
		}

	case NWAError:
		fmt.Printf("%#v\n", v)
	default:
		fmt.Printf("%#v\n", v)
	}

	start := b.start()
	reset := b.reset()
	split := b.split(splitIndex)

	return nwaSummary{
		Start: start,
		Reset: reset,
		Split: split,
	}, nil
}

type nwaSummary struct {
	Start bool
	Reset bool
	Split bool
}

// Checks conditions and returns start state
func (b *NWASplitter) start() bool {
	fmt.Printf("Checking start state\n")
	for _, p := range b.startConditions {
		startState := true
		var tempstate bool

		for _, q := range p.memory {
			watcher := findMemoryWatcher(b.nwaMemory, q.memoryEntryName)
			tempstate = q.result(q.compareType, watcher.priorValue, watcher.currentValue, q.expectedValue)
			startState = startState && tempstate
		}
		if startState {
			fmt.Printf("Start: %#v\n", p.Name)
			return true
		}
	}
	return false
}

// Checks conditions and returns reset state
func (b *NWASplitter) reset() bool {
	fmt.Printf("Checking reset state\n")
	for _, p := range b.resetConditions {
		resetState := true
		var tempstate bool

		for _, q := range p.memory {
			watcher := findMemoryWatcher(b.nwaMemory, q.memoryEntryName)
			tempstate = q.result(q.compareType, watcher.priorValue, watcher.currentValue, q.expectedValue)
			resetState = resetState && tempstate
		}
		if resetState {
			fmt.Printf("Reset: %#v\n", p.Name)
			return true
		}
	}
	return false
}

// Checks conditions and returns split state
func (b *NWASplitter) split(split int) bool {
	fmt.Printf("Checking split state\n")
	splitState := true
	var tempstate bool

	for _, q := range b.splitConditions[split].memory {
		watcher := findMemoryWatcher(b.nwaMemory, q.memoryEntryName)
		tempstate = q.result(q.compareType, watcher.priorValue, watcher.currentValue, q.expectedValue)
		splitState = splitState && tempstate
	}
	if splitState {
		fmt.Printf("Split: %#v\n", b.splitConditions[split].Name)
		return true
	}
	return false
}

// private
func findMemoryWatcher(memInfo []memoryWatcher, targetWatcher string) *memoryWatcher {
	for _, watcher := range memInfo {
		if watcher.name == targetWatcher {
			return &watcher
		}
	}
	return nil
}

// convert hex string to int
func hexToInt(hex string) *int {
	num, err := strconv.ParseUint(hex, 0, 64)
	if err != nil {
		log.Printf("Failed to convert string to integer: %v", err)
		return nil
	}
	integer := int(num)
	return &integer
}

func compare(input string, prior *int, current *int, expected *int) bool {
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
	case "pbse":
		if (expected == nil) || (prior == nil) {
			return false
		} else {
			return (*prior & *expected) != 0
		}
	case "cbse":
		if (expected == nil) || (current == nil) {
			return false
		} else {
			return (*current & *expected) != 0
		}
	case "pbue":
		if (expected == nil) || (prior == nil) {
			return false
		} else {
			return (*prior & *expected) == 0
		}
	case "cbue":
		if (expected == nil) || (current == nil) {
			return false
		} else {
			return (*current & *expected) == 0
		}
	// case "pboe":
	// 	if (expected == nil) || (prior == nil) {
	// 		return false
	// 	} else {
	// 		return (*prior | *expected) != 0
	// 	}
	// case "cboe":
	// 	if (expected == nil) || (current == nil) {
	// 		return false
	// 	} else {
	// 		return (*current | *expected) != 0
	// 	}
	// case "pbne":
	// 	if (expected == nil) || (prior == nil) {
	// 		return false
	// 	} else {
	// 		return (*prior ^ *expected) != 0
	// 	}
	// case "cbne":
	// 	if (expected == nil) || (current == nil) {
	// 		return false
	// 	} else {
	// 		return (*current ^ *expected) != 0
	// 	}
	default:
		return false
	}
}
