//go:build linux && x11

package x11

/*
#cgo pkg-config: x11 xi
#include "provider_x11.h"
*/
import "C"
import (
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/zellydev-games/opensplit/keyinfo"
	"github.com/zellydev-games/opensplit/logger"
)

const logModule = "hotkeys"

type Manager struct {
	callback   func(data keyinfo.KeyData)
	mu         sync.Mutex
	started    bool
	lastUpdate time.Time
}

// SetupHotkeys implements the HotKeyProvider interface to deliver a manager to the caller
func SetupHotkeys() *Manager {
	return new(Manager)
}

// StartHook starts the low level hotkey listener and calls the provided callback when a keypress event is detected
//
// The underlying low level listener needs to deliver a keycode and the name of the key
func (x *Manager) StartHook(callback func(data keyinfo.KeyData)) error {
	logger.Info(logModule, "starting x11 hotkey producer hook")
	x.mu.Lock()
	x.callback = callback
	if x.started {
		x.mu.Unlock()
		logger.Debug(logModule, "previously started, updating callback and leaving")
		return nil
	}
	x.started = true
	logger.Debug(logModule, "x11 hotkey producer flagged as started")
	x.mu.Unlock()

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		var ebuf [128]byte

		if rc := C.xi2_open((*C.char)(unsafe.Pointer(&ebuf[0])), C.int(len(ebuf))); rc != 0 {
			message := C.GoString((*C.char)(unsafe.Pointer(&ebuf[0])))
			logger.Error(message)
			x.mu.Lock()
			x.started = false
			x.mu.Unlock()
			return
		}
		logger.Debug(logModule, "x11 display opened and raw event selector installed")

		var ev C.xi2_event
		for {
			// Blocking call into C; wakes on key events.
			if C.xi2_next(&ev) != 0 {
				C.xi2_close()
				x.mu.Lock()
				x.started = false
				x.mu.Unlock()
				return
			}

			x.mu.Lock()
			cb := x.callback
			x.mu.Unlock()
			// Dedupe on a timer to ignore events from master/non-master devices (common in VMs)
			if cb != nil && time.Now().Sub(x.lastUpdate) > time.Millisecond*20 {
				cb(keyinfo.KeyData{
					KeyCode:    int(ev.keycode),
					LocaleName: C.GoString(&ev.name[0]),
				})
				x.lastUpdate = time.Now()
			}
		}
	}()
	logger.Info(logModule, "x11 hotkey producer started")
	return nil
}

func (x *Manager) Unhook() error {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.callback = nil
	logger.Debug(logModule, "x11 hotkey producer unhook")
	return nil
}
