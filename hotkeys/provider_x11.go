//go:build linux && x11

package hotkeys

/*
#cgo pkg-config: x11 xi
#include "provider_x11.h"
*/
import "C"
import (
	"errors"
	"unsafe"

	"github.com/zellydev-games/opensplit/keyinfo"
)

type X11Manager struct{}

// SetupHotkeys implements the HotKeyProvider interface to deliver a manager to the caller
func SetupHotkeys() *X11Manager {
	return new(X11Manager)
}

// StartHook starts the low level hotkey listener and calls the provided callback when a keypress event is detected
//
// The underlying low level listener needs to deliver a keycode and the name of the key
func (x *X11Manager) StartHook(callback func(data keyinfo.KeyData)) error {
	var ebuf [128]byte
	if rc := C.xi2_open((*C.char)(unsafe.Pointer(&ebuf[0])), C.int(len(ebuf))); rc != 0 {
		return errors.New(C.GoString((*C.char)(unsafe.Pointer(&ebuf[0]))))
	}

	go func() {
		var ev C.xi2_event

		for {
			// Blocking call into C; wakes on key events.
			if C.xi2_next(&ev) != 0 {
				C.xi2_close()
				return
			}

			callback(keyinfo.KeyData{
				KeyCode:    int(ev.keycode),
				LocaleName: C.GoString(&ev.name[0]),
			})
		}
	}()

	return nil
}

func (x *X11Manager) Unhook() error {
	C.xi2_close()
	return nil
}
