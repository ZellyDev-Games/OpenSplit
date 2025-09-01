//go:build windows

package timer

import (
	"OpenSplit2/timer"
	"context"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

/*
Win32 bits
*/
var (
	user32               = windows.NewLazySystemDLL("user32.dll")
	kernel32             = windows.NewLazySystemDLL("kernel32.dll")
	pSetWindowsHookEx    = user32.NewProc("SetWindowsHookExW")
	pUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	pCallNextHookEx      = user32.NewProc("CallNextHookEx")
	pGetMessage          = user32.NewProc("GetMessageW")
	pTranslateMessage    = user32.NewProc("TranslateMessage")
	pDispatchMessage     = user32.NewProc("DispatchMessageW")
	pPostThreadMessage   = user32.NewProc("PostThreadMessageW")
	pGetCurrentThreadId  = kernel32.NewProc("GetCurrentThreadId")
	procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")
)

type Mod uint8

const (
	ModCtrl  Mod = 1 << 0
	ModAlt   Mod = 1 << 1
	ModShift Mod = 1 << 2
	ModWin   Mod = 1 << 3
)

// Windows VKs we'll query
const (
	VK_SHIFT   = 0x10
	VK_CONTROL = 0x11
	VK_MENU    = 0x12 // Alt
	VK_LWIN    = 0x5B
	VK_RWIN    = 0x5C
)

const (
	WhKeyboardLl = 13
	WmKeydown    = 0x0100
	WmSyskeydown = 0x0104
	WmQuit       = 0x0012

	VkSpace = 0x20
)

type (
	HHOOK   uintptr
	WPARAM  uintptr
	LPARAM  uintptr
	LRESULT uintptr

	MSG struct {
		Hwnd    uintptr
		Message uint32
		WParam  WPARAM
		LParam  LPARAM
		Time    uint32
		Pt      struct{ X, Y int32 }
	}

	// KBDLLHOOKSTRUCT from winuser.h
	KBDLLHOOKSTRUCT struct {
		VkCode      uint32
		ScanCode    uint32
		Flags       uint32
		Time        uint32
		DwExtraInfo uintptr
	}
)

type TimerHotkeys struct {
	ctx      context.Context
	cancel   context.CancelFunc
	Timer    *timer.Service
	hhook    HHOOK
	threadID uint32
}

func (t *TimerHotkeys) Startup(ctx context.Context, tm *timer.Service) {
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.Timer = tm
	go t.runHookThread() // isolated OS thread with message loop
}

func (t *TimerHotkeys) Shutdown() {
	if t.cancel != nil {
		t.cancel()
	}
}

func getAsyncKeyState(vk uintptr) int16 {
	r, _, _ := procGetAsyncKeyState.Call(vk)
	return int16(r)
}

// "Down" if the high bit is set (bit 15). That's the Win32 rule.
func isDown(vk uintptr) bool {
	return (getAsyncKeyState(vk) & int16(0x8000)) != 0
}

func currentMods() Mod {
	var m Mod
	if isDown(VK_CONTROL) {
		m |= ModCtrl
	}
	if isDown(VK_MENU) {
		m |= ModAlt
	}
	if isDown(VK_SHIFT) {
		m |= ModShift
	}
	// Consider either Win key as "Win"
	if isDown(VK_LWIN) || isDown(VK_RWIN) {
		m |= ModWin
	}
	return m
}

func (t *TimerHotkeys) runHookThread() {
	// Hook must live on a single OS thread with a message loop.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save thread id (used to post WM_QUIT on shutdown)
	t.threadID = getCurrentThreadID()

	// Install hook
	hookProc := windows.NewCallback(func(nCode int, wparam WPARAM, lparam LPARAM) LRESULT {
		// nCode < 0 → pass through immediately
		if nCode < 0 {
			return callNextHook(t.hhook, nCode, wparam, lparam)
		}

		msg := uint32(wparam)
		if msg == WmKeydown || msg == WmSyskeydown {
			// Read keyboard data
			kb := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lparam))
			if kb.VkCode == VkSpace {
				// Toggle timer — but DO NOT eat the key.
				if !t.Timer.IsRunning() {
					t.Timer.Start()
				} else {
					t.Timer.Pause()
				}
				// Pass-through: do NOT return non-zero. Fall-through to CallNextHookEx.
			}
		}
		return callNextHook(t.hhook, nCode, wparam, lparam)
	})

	// SetWindowsHookExW(WH_KEYBOARD_LL, hookProc, 0, 0)
	h, _, _ := pSetWindowsHookEx.Call(
		uintptr(WhKeyboardLl),
		hookProc,
		0, // hMod = 0: hook proc in this process
		0, // dwThreadId = 0: global
	)
	if h == 0 {
		// Could log: install failed
		return
	}
	t.hhook = HHOOK(h)
	defer func() {
		if t.hhook != 0 {
			pUnhookWindowsHookEx.Call(uintptr(t.hhook))
			t.hhook = 0
		}
	}()

	// Pump messages until context is cancelled.
	msg := MSG{}
	for {
		select {
		case <-t.ctx.Done():
			// Break the GetMessage loop by posting WM_QUIT to this thread.
			postThreadMessage(t.threadID, WmQuit, 0, 0)
			return
		default:
		}

		r, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		switch int32(r) {
		case -1:
			// GetMessage error → exit
			return
		case 0:
			// WM_QUIT received → exit
			return
		default:
			pTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			pDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
		}
	}
}

/*
Helper wrappers
*/
func callNextHook(h HHOOK, nCode int, w WPARAM, l LPARAM) LRESULT {
	r, _, _ := pCallNextHookEx.Call(uintptr(h), uintptr(nCode), uintptr(w), uintptr(l))
	return LRESULT(r)
}

func postThreadMessage(threadID uint32, msg uint32, w WPARAM, l LPARAM) bool {
	r, _, _ := pPostThreadMessage.Call(uintptr(threadID), uintptr(msg), uintptr(w), uintptr(l))
	return r != 0
}

func getCurrentThreadID() uint32 {
	r, _, _ := pGetCurrentThreadId.Call()
	return uint32(r)
}
