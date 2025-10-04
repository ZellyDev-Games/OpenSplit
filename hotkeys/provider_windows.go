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
	whKeyboardLL = 13
	wmKeyDown    = 0x0100
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
	if int(nCode) < 0 {
		ret, _, _ := callNextHook.Call(uintptr(0), nCode, identifier, kbHookStruct)
		return ret
	}

	// If nCode is 0, this message's parameters contains keyboard information, so it's the one we're looking for
	if nCode == 0 {
		if identifier == wmKeyDown {
			// This is a keydown event, the one we care about
			hookInfo := *(*kbDLLHook)(unsafe.Pointer(kbHookStruct)) //nolint:all
			extended := hookInfo.flags&0x1 == 1
			var lparam uintptr
			var buf = make([]uint16, 64)
			p := unsafe.SliceData(buf)
			lparam |= uintptr(hookInfo.scanCode) << 16
			if extended {
				lparam |= 1 << 24
			}

			nameLen, _, err := getKeyName.Call(lparam, uintptr(unsafe.Pointer(p)), uintptr(len(buf)))
			if nameLen == 0 {
				logger.Error(err.Error())
			}

			localeString := windows.UTF16ToString(buf)
			if w.keyPressedCallback != nil {
				w.keyPressedCallback(keyinfo.KeyData{
					KeyCode:    int(hookInfo.vkCode),
					LocaleName: localeString,
				})
			}
		}

		ret, _, _ := callNextHook.Call(uintptr(0), nCode, identifier, kbHookStruct)
		return ret
	}

	// Calling nextHook is optional when nCode isn't negative, but encouraged, so let's do it.
	ret, _, err := callNextHook.Call(uintptr(0), uintptr(nCode), identifier, kbHookStruct)
	if err != nil {
		logger.Error(err.Error())
	}
	return ret
}
