//go:build linux && x11

package x11

/*
#cgo pkg-config: x11 xi
#include "provider_x11.h"
*/
import "C"

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/keyinfo"
	"github.com/zellydev-games/opensplit/logger"
)

const logModule = "hotkeys"

const (
	x11ShiftMask   = 1 << 0
	x11ControlMask = 1 << 2
	x11AltMask     = 1 << 3
	x11MetaMask    = 1 << 6
)

type Manager struct {
	mu sync.RWMutex

	callback  func(command.Command)
	keyConfig map[command.Command]keyinfo.KeyData

	started bool
	enabled bool

	lastUpdate time.Time

	stopOnce sync.Once
	done     chan struct{}
}

func SetupHotkeys() *Manager {
	return &Manager{
		keyConfig: make(map[command.Command]keyinfo.KeyData),
	}
}

func (x *Manager) Start(ctx context.Context) error {
	x.mu.Lock()

	if x.started {
		x.mu.Unlock()
		return nil
	}

	x.stopOnce = sync.Once{}
	x.started = true
	x.enabled = false
	x.done = make(chan struct{})

	done := x.done

	x.mu.Unlock()

	logger.Info(logModule, "starting x11 global hotkey provider")

	go x.run(ctx, done)

	return nil
}

func (x *Manager) run(
	ctx context.Context,
	done chan struct{},
) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	defer close(done)

	var errbuf [128]C.char

	if rc := C.xi2_open(
		(*C.char)(unsafe.Pointer(&errbuf[0])),
		C.int(len(errbuf)),
	); rc != 0 {
		message := C.GoString(
			(*C.char)(unsafe.Pointer(&errbuf[0])),
		)

		logger.Error(
			logModule,
			"x11 hotkey provider failed to open: "+message,
		)

		x.mu.Lock()
		x.started = false
		x.mu.Unlock()

		return
	}

	logger.Info(
		logModule,
		"x11 global hotkey provider started",
	)

	go func() {
		<-ctx.Done()
		C.xi2_stop()
	}()

	for {
		var event C.xi2_event

		rc := C.xi2_next(&event)

		if rc == 2 {
			break
		}

		if rc != 0 {
			logger.Error(
				logModule,
				"x11 hotkey event reader stopped",
			)
			break
		}

		if C.uint8_t(event._type) != 1 {
			continue
		}

		x.handleEvent(event)
	}

	C.xi2_close()

	x.mu.Lock()
	x.started = false
	x.enabled = false
	x.mu.Unlock()

	logger.Info(
		logModule,
		"x11 global hotkey provider stopped",
	)
}

func (x *Manager) handleEvent(event C.xi2_event) {
	x.mu.RLock()

	enabled := x.enabled
	callback := x.callback

	var matched command.Command
	var found bool

	if enabled && callback != nil {
		keyCode := int(event.keycode)
		localeName := C.GoString(&event.name[0])
		modifiers := x11Modifiers(uint32(event.modifiers))

		for cmd, binding := range x.keyConfig {
			if matchesKey(
				keyCode,
				localeName,
				modifiers,
				binding,
			) {
				matched = cmd
				found = true
				break
			}
		}
	}

	x.mu.RUnlock()

	if !found {
		return
	}

	/*
	 * Raw XInput can report the same physical key through multiple
	 * devices. Keep the existing small debounce behavior.
	 */
	x.mu.Lock()

	now := time.Now()

	if now.Sub(x.lastUpdate) <= 20*time.Millisecond {
		x.mu.Unlock()
		return
	}

	x.lastUpdate = now

	x.mu.Unlock()

	callback(matched)
}

func (x *Manager) Configure(
	keyConfig map[command.Command]keyinfo.KeyData,
) error {
	copied := make(map[command.Command]keyinfo.KeyData, len(keyConfig))

	for cmd, binding := range keyConfig {
		copied[cmd] = keyinfo.NewKeyData(
			binding.KeyCode,
			binding.LocaleName,
			append([]string(nil), binding.Modifiers...),
			append([]string(nil), binding.ModifierLocaleNames...),
		)
	}

	x.mu.Lock()
	x.keyConfig = copied
	x.mu.Unlock()

	logger.Debug(
		logModule,
		fmt.Sprintf(
			"x11 global hotkeys configured: %d bindings",
			len(copied),
		),
	)

	return nil
}

func (x *Manager) Enable() error {
	x.mu.Lock()

	if !x.started {
		x.mu.Unlock()
		return fmt.Errorf("x11 hotkey provider is not started")
	}

	x.enabled = true

	x.mu.Unlock()

	logger.Info(
		logModule,
		"x11 global hotkeys enabled",
	)

	return nil
}

func (x *Manager) Disable() error {
	x.mu.Lock()
	x.enabled = false
	x.mu.Unlock()

	logger.Info(
		logModule,
		"x11 global hotkeys disabled",
	)

	return nil
}

func (x *Manager) SetCommandCallback(
	callback func(command.Command),
) {
	x.mu.Lock()
	x.callback = callback
	x.mu.Unlock()
}

func (x *Manager) Close() error {
	x.mu.Lock()

	if !x.started {
		x.mu.Unlock()
		return nil
	}

	done := x.done
	x.enabled = false

	x.mu.Unlock()

	logger.Info(
		logModule,
		"closing x11 global hotkey provider",
	)

	x.stopOnce.Do(func() {
		C.xi2_stop()
	})

	<-done

	return nil
}

func x11Modifiers(mask uint32) []string {
	modifiers := make([]string, 0, 4)

	if mask&x11ControlMask != 0 {
		modifiers = append(modifiers, "CTRL")
	}

	if mask&x11AltMask != 0 {
		modifiers = append(modifiers, "ALT")
	}

	if mask&x11ShiftMask != 0 {
		modifiers = append(modifiers, "SHIFT")
	}

	if mask&x11MetaMask != 0 {
		modifiers = append(modifiers, "META")
	}

	return modifiers
}

func matchesKey(
	keyCode int,
	localeName string,
	modifiers []string,
	binding keyinfo.KeyData,
) bool {
	if keyCode != binding.KeyCode {
		return false
	}

	/*
	 * KeyCode is the authoritative identity, matching the frontend
	 * configuration behavior. LocaleName is intentionally not used
	 * for matching because the same physical key can have different
	 * names depending on the keyboard layout.
	 */
	_ = localeName

	required := normalizeModifiers(binding.Modifiers)

	if len(required) != len(modifiers) {
		return false
	}

	for _, requiredModifier := range required {
		found := false

		for _, modifier := range modifiers {
			if requiredModifier == modifier {
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

func normalizeModifiers(modifiers []string) []string {
	result := make([]string, 0, len(modifiers))

	for _, modifier := range modifiers {
		switch strings.ToUpper(modifier) {
		case "CONTROL":
			result = append(result, "CTRL")

		case "COMMAND":
			result = append(result, "META")

		default:
			result = append(result, strings.ToUpper(modifier))
		}
	}

	return result
}
