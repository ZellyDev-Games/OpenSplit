package qusb2snes

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type conditionList struct {
	Name   string
	memory []element
}

type element struct {
	memoryEntryName string
	expectedValue   *uint16
	compareType     string
	result          compareFunc
}

type compareFunc func(input string, prior *uint16, current *uint16, expected *uint16) bool

type memoryWatcher struct {
	// name    string
	address uint32
	current *uint16
	old     *uint16
	size    int
}

type SNESState struct {
	vars            map[string]*memoryWatcher
	data            []byte
	startConditions []conditionList
	resetConditions []conditionList
	splitConditions []conditionList
	// doExtraUpdate   bool
	// pickedUpHundredthMissile bool
	// pickedUpSporeSpawnSuper bool
	// latencySamples []uint128
	// mu                       sync.Mutex
}

// const NUM_LATENCY_SAMPLES = 10

// type uint128 struct {
// 	hi uint64
// 	lo uint64
// }

// func (a uint128) Add(b uint128) uint128 {
// 	lo := a.lo + b.lo
// 	hi := a.hi + b.hi
// 	if lo < a.lo {
// 		hi++
// 	}
// 	return uint128{hi: hi, lo: lo}
// }

// func (a uint128) Sub(b uint128) uint128 {
// 	lo := a.lo - b.lo
// 	hi := a.hi - b.hi
// 	if a.lo < b.lo {
// 		hi--
// 	}
// 	return uint128{hi: hi, lo: lo}
// }

// func (a uint128) ToFloat64() float64 {
// 	return float64(a.hi)*math.Pow(2, 64) + float64(a.lo)
// }

// func uint128FromInt(i int64) uint128 {
// 	if i < 0 {
// 		return uint128{hi: math.MaxUint64, lo: uint64(i)}
// 	}
// 	return uint128{hi: 0, lo: uint64(i)}
// }

// func averageUint128Slice(arr []uint128) float64 {
// 	var sum uint128
// 	for _, v := range arr {
// 		sum = sum.Add(v)
// 	}
// 	return sum.ToFloat64() / float64(len(arr))
// }

// type TimeSpan struct {
// 	seconds float64
// }

// func (t TimeSpan) Seconds() float64 {
// 	return t.seconds
// }

// func TimeSpanFromSeconds(seconds float64) TimeSpan {
// 	return TimeSpan{seconds: seconds}
// }

// func (s *SNESState) gametimeToSeconds() TimeSpan {
// 	hours := float64(s.vars["igtHours"].current)
// 	minutes := float64(s.vars["igtMinutes"].current)
// 	seconds := float64(s.vars["igtSeconds"].current)

// 	totalSeconds := hours*3600 + minutes*60 + seconds
// 	return TimeSpanFromSeconds(totalSeconds)
// }

// func (a *QUSB2SNESAutoSplitter) GametimeToSeconds() *TimeSpan {
// 	t := a.snes.gametimeToSeconds()
// 	return &t
// }

func (s *SNESState) split(split int) bool {
	splitState := true
	var tempstate bool

	for _, q := range s.splitConditions[split].memory {
		tempstate = q.result(q.compareType, s.vars[q.memoryEntryName].old, s.vars[q.memoryEntryName].current, q.expectedValue)
		splitState = splitState && tempstate
	}
	if splitState {
		fmt.Printf("Split: %#v\n", s.splitConditions[split].Name)
		return true
	}
	return false
}

func (s *SNESState) start() bool {
	for _, p := range s.startConditions {
		startState := true
		var tempstate bool

		for _, q := range p.memory {
			tempstate = q.result(q.compareType, s.vars[q.memoryEntryName].old, s.vars[q.memoryEntryName].current, q.expectedValue)
			startState = startState && tempstate
		}
		if startState {
			fmt.Printf("Start: %#v\n", p.Name)
			return true
		}
	}
	return false
}

func (s *SNESState) reset() bool {
	for _, p := range s.resetConditions {
		resetState := true
		var tempstate bool

		for _, q := range p.memory {
			tempstate = q.result(q.compareType, s.vars[q.memoryEntryName].old, s.vars[q.memoryEntryName].current, q.expectedValue)
			resetState = resetState && tempstate
		}
		if resetState {
			fmt.Printf("Reset: %#v\n", p.Name)
			return true
		}
	}
	return false
}

func newSNESState(memData []string, startConditionImport []string, resetConditionImport []string, splitConditionImport []string) *SNESState {
	data := make([]byte, 0x10000)
	vars := map[string]*memoryWatcher{}

	delimiter1 := "="
	delimiter2 := "≠"
	delimiter3 := "<"
	delimiter4 := ">"
	delimiter5 := "&"
	delimiter6 := "|"
	delimiter7 := "^"
	compareStringCurrent := "current"
	compareStringPrior := "prior"
	var startConditions []conditionList
	var resetConditions []conditionList
	var splitConditions []conditionList

	// fill vars map
	for _, p := range memData {
		mem := strings.Split(p, ",")
		size, _ := strconv.Atoi(mem[2])
		tempHex := *hexToInt(mem[1])
		temp32int := uint32(tempHex)
		vars[mem[0]] = newMemoryWatcher(temp32int, size)
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
		startConditions = append(startConditions, condition)
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
		resetConditions = append(resetConditions, condition)
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
		splitConditions = append(splitConditions, condition)
	}

	return &SNESState{
		// doExtraUpdate:   true,
		data:            data,
		startConditions: startConditions,
		resetConditions: resetConditions,
		splitConditions: splitConditions,
		// latencySamples: make([]uint128, 0),
		// pickedUpHundredthMissile: false,
		// pickedUpSporeSpawnSuper:  false,
		vars: vars,
	}
}

func newMemoryWatcher(address uint32, size int) *memoryWatcher {
	return &memoryWatcher{
		address: address,
		current: new(uint16),
		old:     new(uint16),
		size:    size,
	}
}

func (mw *memoryWatcher) updateValue(memory []byte) {
	*mw.old = *mw.current
	switch mw.size {
	case 1:
		*mw.current = uint16(memory[mw.address])
	case 2:
		addr := mw.address
		*mw.current = uint16(memory[addr]) | uint16(memory[addr+1])<<8
	}
}

func (s *SNESState) update() {
	// s.mu.Lock()
	// defer s.mu.Unlock()
	for _, watcher := range s.vars {
		// if s.doExtraUpdate {
		// watcher.updateValue(s.data)
		// s.doExtraUpdate = false
		// }
		watcher.updateValue(s.data)
	}
}

type SNESSummary struct {
	// LatencyAverage float64
	// LatencyStddev  float64
	Start bool
	Reset bool
	Split bool
}

type QUSB2SNESAutoSplitter struct {
	snes *SNESState
	// settings     *sync.RWMutex
	// settingsData *Settings
}

func NewQUSB2SNESAutoSplitter(memData []string, startConditionImport []string, resetConditionImport []string, splitConditionImport []string /*settings *sync.RWMutex, settingsData *Settings*/) *QUSB2SNESAutoSplitter {
	return &QUSB2SNESAutoSplitter{
		snes: newSNESState(memData, startConditionImport, resetConditionImport, splitConditionImport),
		// settings:     settings,
		// settingsData: settingsData,
	}
}

func (a *QUSB2SNESAutoSplitter) Update(client SyncClient, splitNum int) (*SNESSummary, error) {
	addresses := [][2]int{}

	for _, watcher := range a.snes.vars {
		fullAddress := int(watcher.address | 0xF50000)

		newRow := []int{fullAddress, int(watcher.size)}
		addresses = append(addresses, [2]int(newRow))
	}

	snesData, err := client.getAddresses(addresses)
	if err != nil {
		return nil, err
	}

	for index, row := range addresses {
		copy(a.snes.data[(row[0]^0xF50000):(row[0]^0xF50000)+row[1]], snesData[index])
	}
	a.snes.update()

	start := a.snes.start()
	reset := a.snes.reset()
	split := a.snes.split(splitNum)

	// elapsed := time.Since(startTime).Milliseconds()

	// if len(s.latencySamples) == NUM_LATENCY_SAMPLES {
	// s.latencySamples = s.latencySamples[1:]
	// }
	// s.latencySamples = append(s.latencySamples, uint128FromInt(elapsed))

	// averageLatency := averageUint128Slice(s.latencySamples)

	// var sdevSum float64
	// for _, x := range s.latencySamples {
	// diff := x.ToFloat64() - averageLatency
	// sdevSum += diff * diff
	// }
	// stddev := math.Sqrt(sdevSum / float64(len(s.latencySamples)-1))

	return &SNESSummary{
		// LatencyAverage: averageLatency,
		// LatencyStddev:  stddev,
		Start: start,
		Reset: reset,
		Split: split,
	}, nil
}

func (a *QUSB2SNESAutoSplitter) ResetGameTracking() {
	// a.snes = newSNESState()
	clear(a.snes.data[:])
	for _, watcher := range a.snes.vars {
		*watcher.current = 0
		*watcher.old = 0
	}
	// a.snes.doExtraUpdate = true
}

// convert hex string to int
func hexToInt(hex string) *uint16 {
	num, err := strconv.ParseUint(hex, 0, 64)
	if err != nil {
		log.Printf("Failed to convert string to integer: %v", err)
		return nil
	}
	integer := uint16(num)
	return &integer
}

func compare(input string, prior *uint16, current *uint16, expected *uint16) bool {
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
	case "pbae":
		if (expected == nil) || (prior == nil) {
			return false
		} else {
			return (*prior & *expected) != 0
		}
	case "cbae":
		if (expected == nil) || (current == nil) {
			return false
		} else {
			return (*current & *expected) != 0
		}
	case "pboe":
		if (expected == nil) || (prior == nil) {
			return false
		} else {
			return (*prior | *expected) != 0
		}
	case "cboe":
		if (expected == nil) || (current == nil) {
			return false
		} else {
			return (*current | *expected) != 0
		}
	case "pbne":
		if (expected == nil) || (prior == nil) {
			return false
		} else {
			return (*prior ^ *expected) != 0
		}
	case "cbne":
		if (expected == nil) || (current == nil) {
			return false
		} else {
			return (*current ^ *expected) != 0
		}
	default:
		return false
	}
}
