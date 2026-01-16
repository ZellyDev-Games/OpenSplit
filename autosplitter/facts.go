package autosplitter

import (
	"fmt"
	"sync"

	"github.com/zellydev-games/opensplit/logger"
)

type ValueType byte

const (
	ValueSigned ValueType = iota
	ValueUnsigned
	ValueBoolean
)

type ValueTable struct {
	m              sync.RWMutex
	signedValues   map[string]int64
	unsignedValues map[string]uint64
	booleanValues  map[string]bool
}

func NewValueTable() *ValueTable {
	return &ValueTable{
		signedValues:   map[string]int64{},
		unsignedValues: map[string]uint64{},
		booleanValues:  map[string]bool{},
	}
}

func (t *ValueTable) SetValue(id string, val int64, valType ValueType) bool {
	t.m.Lock()
	defer t.m.Unlock()

	switch valType {
	case ValueSigned:
		t.signedValues[id] = val
	case ValueUnsigned:
		if val < 0 {
			logger.Error(fmt.Sprintf("attempt to set negative unsigned value fact for key %s", id))
			return false
		}
		t.unsignedValues[id] = uint64(val)
	case ValueBoolean:
		t.booleanValues[id] = val != 0
	default:
		logger.Error(fmt.Sprintf("attempt to set invalid value type for key %s", id))
		return false
	}

	return true
}

func (t *ValueTable) GetUnsignedValue(id string) (uint64, bool) {
	t.m.RLock()
	defer t.m.RUnlock()
	val, ok := t.unsignedValues[id]
	return val, ok
}

func (t *ValueTable) GetSignedValue(id string) (int64, bool) {
	t.m.RLock()
	defer t.m.RUnlock()
	val, ok := t.signedValues[id]
	return val, ok
}

func (t *ValueTable) GetBoolValue(id string) (bool, bool) {
	t.m.RLock()
	defer t.m.RUnlock()
	val, ok := t.booleanValues[id]
	return val, ok
}

func (t *ValueTable) Snapshot() (map[string]int64, map[string]uint64, map[string]bool) {
	t.m.RLock()
	defer t.m.RUnlock()

	sv := make(map[string]int64, len(t.signedValues))
	uv := make(map[string]uint64, len(t.unsignedValues))
	bv := make(map[string]bool, len(t.booleanValues))

	for k, v := range t.signedValues {
		sv[k] = v
	}

	for k, v := range t.unsignedValues {
		uv[k] = v
	}

	for k, v := range t.booleanValues {
		bv[k] = v
	}

	return sv, uv, bv
}
