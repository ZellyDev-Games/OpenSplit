//go:build windows

package hotkeys

import (
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"github.com/zellydev-games/opensplit/keyinfo"
	"github.com/zellydev-games/opensplit/logger"

	"golang.org/x/sys/windows"
)

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	setWindowsHook    = user32.NewProc("SetWindowsHookExW")
	unhookWindowsHook = user32.NewProc("UnhookWindowsHookEx")
	callNextHook      = user32.NewProc("CallNextHookEx")
	getMessage        = user32.NewProc("GetMessageW")
	getKeyName        = user32.NewProc("GetKeyNameTextW")
)

const (
	vkLShift   = 0xA0
	vkRShift   = 0xA1
	vkLControl = 0xA2
	vkRControl = 0xA3
	vkLMenu    = 0xA4
	vkRMenu    = 0xA5
)

const (
	whKeyboardLL = 13
	wmKeyDown    = 0x0100
	wmKeyUp      = 0x0101
	wmSysKeyDown = 0x0104
	wmSysKeyUp   = 0x0105
)

type kbDLLHook struct {
	vkCode    uint32
	scanCode  uint32
	flags     uint32
	time      uint32
	extraInfo uintptr
}

type threadMessage struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	point   point
	private uint32
}

type point struct {
	x int32
	y int32
}

// WindowsManager implements the HotkeyProvider interface for Windows keypresses
//
// It creates a callback that is invoked with a low-level keyboard hook provided by user32.dll, then reports all
// keypresses to keyChannel where the hotkeys.Service routes it appropriately.  It also features a message pump
// that calls GetMessage provided by user32.dll to inform Windows that our thread is cooperating, and therefore
// eligible to have the callback executed.
type WindowsManager struct {
	hhookHandle        uintptr
	callback           uintptr
	hookThread         windows.Handle
	hooked             bool
	keyPressedCallback func(info keyinfo.KeyData)
	mu                 sync.Mutex
}

// SetupHotkeys implements the HotKeyProvider interface to deliver a manager and channel to the caller
func SetupHotkeys() *WindowsManager {
	return new(WindowsManager)
}

// StartHook converts handleKeyDown into a Windows callback via syscall, and installs it to the locked OS thread with
// setWindowsHook.  It then starts a message pump as required by Windows to inform the OS that this thread is cooperating
// which makes it eligible to have its callback function invoked by the OS.
//
// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setwindowshookexw
// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getmessagew
func (w *WindowsManager) StartHook(callback func(data keyinfo.KeyData)) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.hooked {
		logger.Warn("StartHook() called on hooked manager")
		return nil
	}

	w.keyPressedCallback = callback

	go func() {
		// Messages are sent to the thread that installed the hook, so lock this function down to just that thread
		runtime.LockOSThread()
		w.hookThread = windows.CurrentThread()
		w.callback = syscall.NewCallback(w.handleKeyDown)
		hhook, _, err := setWindowsHook.Call(whKeyboardLL,
			w.callback,
			0,
			0)
		if hhook == 0 {
			logger.Error(err.Error())
			return
		}

		w.hhookHandle = hhook
		logger.Debug(fmt.Sprintf("hook set at address %d", hhook))
		for {
			msg := &threadMessage{}
			ret, _, _ := getMessage.Call(uintptr(unsafe.Pointer(msg)), 0, 0, 0)
			if ret == 0 {
				logger.Debug("WM_QUIT received, quitting message loop")
				err = w.Unhook()
				if err != nil {
					return
				}
				return
			}
		}
	}()

	w.hooked = true
	return nil
}

// Unhook called unhookWindowsHook with the address of our hook handle to inform the OS to stop calling our callback
//
// https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-unhookwindowshookex
func (w *WindowsManager) Unhook() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.hooked {
		logger.Warn("Unhook() called on unhooked manager")
		return nil
	}

	ret, _, err := unhookWindowsHook.Call(w.hhookHandle)
	if ret == 0 {
		logger.Error(err.Error())
		return err
	}

	logger.Debug(fmt.Sprintf("hook removed at address %d", w.hhookHandle))
	w.hooked = false
	return nil
}

// handleKeyDown is called by the OS after StartHook installs it. The callback receives nCode, lparam, and wparam as
// defined by the Win32 API: https://learn.microsoft.com/en-us/windows/win32/winmsg/lowlevelkeyboardproc
func (w *WindowsManager) handleKeyDown(nCode uintptr, identifier uintptr, kbHookStruct uintptr) uintptr {
	// If nCode is less than zero we're obligated to pass the message along
	if int32(nCode) < 0 {
		ret, _, _ := callNextHook.Call(0, nCode, identifier, kbHookStruct)
		return ret
	}

	if isKeyEvent(identifier) {
		// Process modifiers first
		hookInfo := *(*kbDLLHook)(unsafe.Pointer(kbHookStruct)) //nolint:all
		vk := hookInfo.vkCode

		extended := hookInfo.flags&0x1 == 1
		var lparam uintptr
		buf := make([]uint16, 64)
		p := unsafe.SliceData(buf)

		lparam |= uintptr(hookInfo.scanCode) << 16
		if extended {
			lparam |= 1 << 24
		}

		modifierState.mu.Lock()
		switch identifier {
		case wmKeyDown, wmSysKeyDown:
			if isModifierKey(vk) {
				modifierState.m[vk] = true
			}
		case wmKeyUp, wmSysKeyUp:
			if isModifierKey(vk) {
				modifierState.m[vk] = false
			}
		}
		modifierState.mu.Unlock()

		if identifier == wmKeyDown || identifier == wmSysKeyDown {
			if !isModifierKey(hookInfo.vkCode) {
				nameLen, _, err := getKeyName.Call(
					lparam,
					uintptr(unsafe.Pointer(p)),
					uintptr(len(buf)),
				)
				if nameLen == 0 {
					logger.Error(err.Error())
				}

				localeString := windows.UTF16ToString(buf)

				modifierState.mu.Lock()
				modifiers := make([]int, 0, len(modifierState.m))
				for code, state := range modifierState.m {
					if state {
						modifiers = append(modifiers, int(code))
					}
				}
				modifierLocaleNames := make([]string, 0, len(modifiers))
				for _, vkInt := range modifiers {
					if name := w.ModCodeToString(vkInt); name != "" {
						modifierLocaleNames = append(modifierLocaleNames, name)
					}
				}
				modifierState.mu.Unlock()

				fmt.Println(localeString)
				fmt.Println(modifierState.m)

				if w.keyPressedCallback != nil {
					w.keyPressedCallback(
						keyinfo.NewKeyData(
							int(hookInfo.vkCode),
							localeString,
							modifiers,
							modifierLocaleNames,
						),
					)
				}
				resetModifiers()
			}
		}
	}

	ret, _, _ := callNextHook.Call(0, nCode, identifier, kbHookStruct)
	return ret
}

func (w *WindowsManager) ModCodeToString(code int) string {
	switch code {
	case vkLShift:
		return "Left Shift"
	case vkRShift:
		return "Right Shift"
	case vkLControl:
		return "Left Control"
	case vkRControl:
		return "Right Control"
	case vkLMenu:
		return "Left Alt"
	case vkRMenu:
		return "Right Alt"
	default:
		return ""
	}
}

func isModifierKey(vk uint32) bool {
	switch vk {
	case vkLShift, vkRShift,
		vkLControl, vkRControl,
		vkLMenu, vkRMenu:
		return true
	default:
		return false
	}
}

var modifierState = struct {
	mu sync.Mutex
	m  map[uint32]bool
}{
	m: map[uint32]bool{
		vkLControl: false,
		vkRControl: false,
		vkLShift:   false,
		vkRShift:   false,
		vkLMenu:    false,
		vkRMenu:    false,
	},
}

func isKeyEvent(identifier uintptr) bool {
	return identifier == wmKeyDown ||
		identifier == wmKeyUp ||
		identifier == wmSysKeyUp ||
		identifier == wmSysKeyDown
}

func resetModifiers() {
	modifierState.mu.Lock()
	defer modifierState.mu.Unlock()
	modifierState.m = map[uint32]bool{
		vkLControl: false,
		vkRControl: false,
		vkLShift:   false,
		vkRShift:   false,
		vkLMenu:    false,
		vkRMenu:    false,
	}
}
