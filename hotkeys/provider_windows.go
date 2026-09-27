//go:build windows

package hotkeys

import (
	"context"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/keyinfo"
	"github.com/zellydev-games/opensplit/logger"

	"golang.org/x/sys/windows"
)

const logModule = "hotkeys"

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

// WindowsManager implements GlobalHotkeyProvider for Windows.
//
// The Windows low-level keyboard hook is used only while global hotkeys
// are enabled. Local/focused key handling is performed by the frontend.
type WindowsManager struct {
	mu sync.Mutex

	ctx context.Context

	hhookHandle uintptr
	callback    uintptr
	hookThread  windows.Handle
	hooked      bool

	keyConfig map[command.Command]keyinfo.KeyData

	commandCallback func(command.Command)
}

// SetupHotkeys creates the Windows global hotkey manager.
func SetupHotkeys() *WindowsManager {
	return new(WindowsManager)
}

// Start initializes the provider.
func (w *WindowsManager) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.ctx != nil {
		return nil
	}

	w.ctx = ctx

	return nil
}

// Configure stores the current global hotkey configuration.
//
// The actual Windows hook is configured when Enable is called.
func (w *WindowsManager) Configure(
	keyConfig map[command.Command]keyinfo.KeyData,
) error {
	configCopy := make(
		map[command.Command]keyinfo.KeyData,
		len(keyConfig),
	)

	for cmd, data := range keyConfig {
		configCopy[cmd] = data
	}

	w.mu.Lock()
	w.keyConfig = configCopy
	w.mu.Unlock()

	return nil
}

// SetCommandCallback installs the callback used when a configured
// global shortcut is activated.
func (w *WindowsManager) SetCommandCallback(
	callback func(command.Command),
) {
	w.mu.Lock()
	w.commandCallback = callback
	w.mu.Unlock()
}

// Enable installs the Windows low-level keyboard hook.
func (w *WindowsManager) Enable() error {
	w.mu.Lock()

	if w.hooked {
		w.mu.Unlock()
		return nil
	}

	w.hooked = true

	w.mu.Unlock()

	go w.runHook()

	return nil
}

// Disable removes the Windows low-level keyboard hook.
func (w *WindowsManager) Disable() error {
	w.mu.Lock()

	if !w.hooked {
		w.mu.Unlock()
		return nil
	}

	handle := w.hhookHandle
	w.hooked = false

	w.mu.Unlock()

	if handle == 0 {
		return nil
	}

	ret, _, err := unhookWindowsHook.Call(handle)
	if ret == 0 {
		logger.Error(logModule, err.Error())
		return err
	}

	logger.Debugf(
		logModule,
		"Windows global keyboard hook removed at address %d",
		handle,
	)

	return nil
}

// Close shuts down the global hotkey provider.
func (w *WindowsManager) Close() error {
	return w.Disable()
}

// runHook installs the low-level keyboard hook and services its
// required Windows message loop.
func (w *WindowsManager) runHook() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	w.mu.Lock()

	if !w.hooked {
		w.mu.Unlock()
		return
	}

	w.callback = syscall.NewCallback(w.handleKeyDown)

	callback := w.callback

	w.mu.Unlock()

	hhook, _, err := setWindowsHook.Call(
		whKeyboardLL,
		callback,
		0,
		0,
	)

	if hhook == 0 {
		w.mu.Lock()
		w.hooked = false
		w.callback = 0
		w.mu.Unlock()

		if err != nil {
			logger.Error(logModule, err.Error())
		}

		return
	}

	w.mu.Lock()
	w.hhookHandle = hhook
	w.hookThread = windows.CurrentThread()
	w.mu.Unlock()

	logger.Debugf(
		logModule,
		"Windows global keyboard hook installed at address %d",
		hhook,
	)

	for {
		msg := &threadMessage{}

		ret, _, _ := getMessage.Call(
			uintptr(unsafe.Pointer(msg)),
			0,
			0,
			0,
		)

		if ret == 0 {
			logger.Debug(
				logModule,
				"WM_QUIT received, stopping Windows global hotkey loop",
			)
			break
		}

		if ret == ^uintptr(0) {
			logger.Error(
				logModule,
				"GetMessageW failed",
			)
			break
		}

		w.mu.Lock()
		active := w.hooked
		w.mu.Unlock()

		if !active {
			break
		}
	}

	// The hook may already have been removed by Disable().
	w.mu.Lock()

	handle := w.hhookHandle
	w.hhookHandle = 0
	w.callback = 0
	w.hookThread = 0
	w.mu.Unlock()

	if handle != 0 {
		_, _, _ = unhookWindowsHook.Call(handle)
	}
}

// handleKeyDown is called by the Windows low-level keyboard hook.
func (w *WindowsManager) handleKeyDown(
	nCode uintptr,
	identifier uintptr,
	kbHookStruct uintptr,
) uintptr {
	if int32(nCode) < 0 {
		return w.callNext(nCode, identifier, kbHookStruct)
	}

	if !isKeyEvent(identifier) {
		return w.callNext(nCode, identifier, kbHookStruct)
	}

	hookInfo := *(*kbDLLHook)(unsafe.Pointer(kbHookStruct)) //nolint:all

	vk := hookInfo.vkCode

	// Keep modifier state synchronized with Windows key events.
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

	// Only key-down events can activate a command.
	if identifier != wmKeyDown && identifier != wmSysKeyDown {
		return w.callNext(nCode, identifier, kbHookStruct)
	}

	// Modifier keys themselves do not activate configured commands.
	if isModifierKey(vk) {
		return w.callNext(nCode, identifier, kbHookStruct)
	}

	localeName := w.keyName(hookInfo)

	modifiers, modifierLocaleNames := currentModifiers()

	data := keyinfo.NewKeyData(
		int(vk),
		localeName,
		modifiers,
		modifierLocaleNames,
	)

	if cmd, ok := w.matchCommand(data); ok {
		w.mu.Lock()
		callback := w.commandCallback
		w.mu.Unlock()

		if callback != nil {
			logger.Debugf(
				logModule,
				"global Windows hotkey activated: command=%d",
				cmd,
			)

			callback(cmd)
		}
	}

	return w.callNext(nCode, identifier, kbHookStruct)
}

func (w *WindowsManager) callNext(
	nCode uintptr,
	identifier uintptr,
	kbHookStruct uintptr,
) uintptr {
	ret, _, _ := callNextHook.Call(
		0,
		nCode,
		identifier,
		kbHookStruct,
	)

	return ret
}

// keyName resolves the Windows localized key name from the keyboard
// scan code.
func (w *WindowsManager) keyName(
	hookInfo kbDLLHook,
) string {
	extended := hookInfo.flags&0x1 == 1

	var lparam uintptr
	lparam |= uintptr(hookInfo.scanCode) << 16

	if extended {
		lparam |= 1 << 24
	}

	buf := make([]uint16, 64)

	nameLen, _, err := getKeyName.Call(
		lparam,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)

	if nameLen == 0 {
		if err != nil {
			logger.Error(logModule, err.Error())
		}

		return ""
	}

	return windows.UTF16ToString(buf)
}

// matchCommand finds the configured command corresponding to a key event.
func (w *WindowsManager) matchCommand(
	data keyinfo.KeyData,
) (command.Command, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for cmd, binding := range w.keyConfig {
		if keyDataMatches(data, binding) {
			return cmd, true
		}
	}

	return 0, false
}

func keyDataMatches(
	event keyinfo.KeyData,
	binding keyinfo.KeyData,
) bool {
	if event.KeyCode != binding.KeyCode {
		return false
	}

	if len(event.Modifiers) != len(binding.Modifiers) {
		return false
	}

	for _, required := range binding.Modifiers {
		found := false

		for _, actual := range event.Modifiers {
			if required == actual {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func currentModifiers() (
	[]string,
	[]string,
) {
	modifierState.mu.Lock()
	defer modifierState.mu.Unlock()

	modifiers := make([]string, 0, len(modifierState.m))
	modifierLocaleNames := make([]string, 0, len(modifierState.m))

	for code, state := range modifierState.m {
		if !state {
			continue
		}

		name := modifierCodeToName(int(code))
		if name != "" {
			modifiers = append(modifiers, name)
		}

		localeName := modifierCodeToLocaleName(int(code))
		if localeName != "" {
			modifierLocaleNames = append(
				modifierLocaleNames,
				localeName,
			)
		}
	}

	return modifiers, modifierLocaleNames
}

func modifierCodeToName(code int) string {
	switch code {
	case vkLShift, vkRShift:
		return "SHIFT"

	case vkLControl, vkRControl:
		return "CTRL"

	case vkLMenu, vkRMenu:
		return "ALT"

	default:
		return ""
	}
}

func modifierCodeToLocaleName(code int) string {
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
	case vkLShift,
		vkRShift,
		vkLControl,
		vkRControl,
		vkLMenu,
		vkRMenu:
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
		identifier == wmSysKeyDown ||
		identifier == wmSysKeyUp
}
