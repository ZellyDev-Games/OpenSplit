//go:build windows

package hotkeys

import (
	"context"
	"fmt"
	"runtime"
	"syscall"
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
	procGetKeyNameTextW  = user32.NewProc("GetKeyNameTextW")
	procMapVirtualKeyW   = user32.NewProc("MapVirtualKeyW")
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
	WhKeyboardLl  = 13
	WmKeydown     = 0x0100
	WmSyskeydown  = 0x0104
	WmQuit        = 0x0012
	MapvkVkToVsc  = 0
	LlkhfExtended = 0x01
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

func getAsyncKeyState(vk uintptr) uint16 {
	r, _, _ := procGetAsyncKeyState.Call(vk)
	return uint16(r)
}

// High bit set => key is currently down (Win32 rule for GetAsyncKeyState)
func isDown(vk uintptr) bool {
	return (getAsyncKeyState(vk) & 0x8000) != 0
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
	if isDown(VK_LWIN) || isDown(VK_RWIN) {
		m |= ModWin
	}
	return m
}

func modsString(m Mod) string {
	out := ""
	if m&ModCtrl != 0 {
		out += "Ctrl+"
	}
	if m&ModAlt != 0 {
		out += "Alt+"
	}
	if m&ModShift != 0 {
		out += "Shift+"
	}
	if m&ModWin != 0 {
		out += "Win+"
	}
	return out
}

func keyNameFromVK(vk uint32, extended bool) string {
	sc, _, _ := procMapVirtualKeyW.Call(uintptr(vk), MapvkVkToVsc)

	// 2) Build LPARAM for GetKeyNameText:
	//    bits 16..23 = scan code; bit 24 = extended key flag
	lparam := uintptr(sc << 16)
	if extended {
		lparam |= 1 << 24
	}

	var buf [64]uint16
	ret, _, _ := procGetKeyNameTextW.Call(
		lparam,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if ret == 0 {
		// Fallback: VK code
		return fmt.Sprintf("VK_%X", vk)
	}
	return syscall.UTF16ToString(buf[:])
}

type KeyboardListener struct {
	ctx      context.Context
	cancel   context.CancelFunc
	hhook    HHOOK
	threadID uint32
}

func NewKeyboardListener(ctx context.Context) *KeyboardListener {
	if ctx == nil {
		ctx = context.Background()
	}
	return &KeyboardListener{ctx: ctx}
}

// Enable installs the LL keyboard hook and starts the message loop on a dedicated thread.
func (l *KeyboardListener) Enable() {
	// If already enabled, no-op.
	if l.cancel != nil {
		return
	}
	l.ctx, l.cancel = context.WithCancel(l.ctx)
	go l.runHookThread(l.ctx)
}

// Disable uninstalls the hook by cancelling the context.
// This triggers ctx.Done() → posts WM_QUIT → exits loop → UnhookWindowsHookEx.
func (l *KeyboardListener) Disable() {
	if l.cancel != nil {
		l.cancel() // 👈 this is how you trigger Done()
		l.cancel = nil
	}
}

func (k *KeyboardListener) Shutdown() {
	if k.cancel != nil {
		k.cancel()
	}
}

func (k *KeyboardListener) runHookThread(ctx context.Context) {
	// Hook must live on a single OS thread with a message loop.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	k.threadID = getCurrentThreadID()

	// Hook procedure
	hookProc := windows.NewCallback(func(nCode int, wparam WPARAM, lparam LPARAM) LRESULT {
		if nCode < 0 {
			return callNextHook(k.hhook, nCode, wparam, lparam)
		}

		msg := uint32(wparam)
		if msg == WmKeydown || msg == WmSyskeydown {
			kb := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lparam))

			// Build modifiers at this moment
			mods := currentMods()

			// Determine if it's an "extended" key (affects naming for arrows, numpad, etc.)
			extended := (kb.Flags & LlkhfExtended) != 0

			// Human-friendly key name
			name := keyNameFromVK(kb.VkCode, extended)

			// Print "Ctrl+Alt+K" etc.
			fmt.Printf("%s%s (VK=0x%X, sc=%d, ext=%t)\n", modsString(mods), name, kb.VkCode, kb.ScanCode, extended)
		}
		return callNextHook(k.hhook, nCode, wparam, lparam)
	})

	// Install global low-level keyboard hook
	h, _, _ := pSetWindowsHookEx.Call(
		uintptr(WhKeyboardLl),
		hookProc,
		0, // hMod = 0 (proc in this process)
		0, // dwThreadId = 0 (global on current desktop)
	)
	if h == 0 {
		return // failed to install
	}
	k.hhook = HHOOK(h)
	defer func() {
		if k.hhook != 0 {
			pUnhookWindowsHookEx.Call(uintptr(k.hhook))
			k.hhook = 0
		}
	}()

	// Pump messages until cancelled
	var msg MSG
	for {
		select {
		case <-ctx.Done():
			postThreadMessage(k.threadID, WmQuit, 0, 0)
			return
		default:
		}
		r, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		switch int32(r) {
		case -1:
			return // GetMessage error
		case 0:
			return // WM_QUIT
		default:
			pTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			pDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
		}
	}
}

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
