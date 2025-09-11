//go:build windows

package hotkeys

import (
	"OpenSplit/logger"
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

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

type KbDLLHook struct {
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

type WindowsManager struct {
	hhookHandle uintptr
	callback    uintptr
	hookThread  windows.Handle
	keyChannel  chan KeyInfo
}

func NewWindowsHotkeyManager() (*WindowsManager, chan KeyInfo) {
	manager := new(WindowsManager)
	manager.keyChannel = make(chan KeyInfo)
	return manager, manager.keyChannel
}

func (h *WindowsManager) StartHook() error {
	go func() {
		// Messages are sent to the thread that installed the hook, so lock this function down to just that thread
		runtime.LockOSThread()
		h.hookThread = windows.CurrentThread()

		logger.Debug("setting windows hook")

		h.callback = syscall.NewCallback(h.HandleKeyDown)
		logger.Debug("created keyboard callback")
		hhook, _, err := setWindowsHook.Call(whKeyboardLL,
			h.callback,
			0,
			0)
		if hhook == 0 {
			logger.Error(err.Error())
			return
		}

		h.hhookHandle = hhook
		logger.Debug(fmt.Sprintf("hook set at address %d", hhook))

		logger.Debug("starting message pump")
		for {
			msg := &threadMessage{}
			ret, _, _ := getMessage.Call(uintptr(unsafe.Pointer(msg)), 0, 0, 0)
			if ret == 0 {
				logger.Debug("WM_QUIT received, quitting message loop")
				err = h.Unhook()
				if err != nil {
					return
				}
				return
			}
		}
	}()

	return nil
}

func (h *WindowsManager) Unhook() error {
	ret, _, err := unhookWindowsHook.Call(h.hhookHandle)
	if ret == 0 {
		logger.Error(err.Error())
		return err
	}
	logger.Debug(fmt.Sprintf("hook removed at address %d", h.hhookHandle))
	return nil
}

func (h *WindowsManager) HandleKeyDown(nCode uintptr, identifier uintptr, kbHookStruct uintptr) uintptr {
	// If nCode is less than zero we're obligated to pass the message along
	if int(nCode) < 0 {
		ret, _, _ := callNextHook.Call(uintptr(0), nCode, identifier, kbHookStruct)
		return ret
	}

	// If nCode is 0, this message's parameters contains keyboard information, so it's the one we're looking for
	if nCode == 0 {
		if identifier == wmKeyDown {
			// This is a keydown event, the one we care about
			hookInfo := *(*KbDLLHook)(unsafe.Pointer(kbHookStruct)) //nolint:all
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
			h.keyChannel <- KeyInfo{
				KeyCode:    int(hookInfo.vkCode),
				LocaleName: localeString,
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
func SetupHotkeys() (HotkeyProvider, chan KeyInfo) {
	var hotkeyProvider HotkeyProvider
	var keyInfoChannel chan KeyInfo
	hotkeyProvider, keyInfoChannel = NewWindowsHotkeyManager()
	return hotkeyProvider, keyInfoChannel
}
