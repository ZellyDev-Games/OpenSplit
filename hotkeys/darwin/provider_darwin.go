//go:build darwin

package darwin

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices

#include <stdint.h>
#include <stdlib.h>

void hk_start(void);
void hk_stop(void);
int  hk_wait_next(
	unsigned short* out_keycode,
	char*           out_name,
	unsigned long   out_name_cap,
	unsigned int*   out_modifiers
);
*/
import "C"

import (
	"context"
	"fmt"
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
	darwinShiftMask   = 1 << 17
	darwinControlMask = 1 << 18
	darwinAltMask     = 1 << 19
	darwinMetaMask    = 1 << 20
)

type Manager struct {
	mu sync.RWMutex

	callback  func(command.Command)
	keyConfig map[command.Command]keyinfo.KeyData

	started bool
	enabled bool

	lastUpdate time.Time

	done chan struct{}
}

func SetupHotkeys() *Manager {
	return &Manager{
		keyConfig: make(map[command.Command]keyinfo.KeyData),
	}
}

func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()

	if m.started {
		m.mu.Unlock()
		return nil
	}

	m.started = true
	m.enabled = false
	m.done = make(chan struct{})

	done := m.done

	m.mu.Unlock()

	logger.Info(logModule, "starting macOS global hotkey provider")

	C.hk_start()

	go m.run(ctx, done)

	return nil
}

func (m *Manager) run(ctx context.Context, done chan struct{}) {
	defer close(done)

	stop := make(chan struct{})
	defer close(stop)

	go func() {
		select {
		case <-ctx.Done():
			C.hk_stop()
		case <-stop:
		}
	}()

	for {
		var keycode C.ushort
		var modifiers C.uint

		buf := make([]byte, 64)

		ok := C.hk_wait_next(
			&keycode,
			(*C.char)(unsafe.Pointer(&buf[0])),
			C.ulong(len(buf)),
			&modifiers,
		)

		if ok == 0 {
			return
		}

		name := C.GoString(
			(*C.char)(unsafe.Pointer(&buf[0])),
		)

		m.handleEvent(
			int(keycode),
			name,
			uint32(modifiers),
		)
	}
}

func (m *Manager) handleEvent(
	keyCode int,
	localeName string,
	modifierMask uint32,
) {
	m.mu.RLock()

	if !m.enabled || m.callback == nil {
		m.mu.RUnlock()
		return
	}

	modifiers := darwinModifiers(modifierMask)

	var (
		matched command.Command
		found   bool
	)

	for cmd, binding := range m.keyConfig {
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

	callback := m.callback

	m.mu.RUnlock()

	if !found || callback == nil {
		return
	}

	// Prevent an accidental duplicate dispatch from the event stream.
	m.mu.Lock()

	now := time.Now()

	if now.Sub(m.lastUpdate) <= 20*time.Millisecond {
		m.mu.Unlock()
		return
	}

	m.lastUpdate = now

	m.mu.Unlock()

	callback(matched)
}

func (m *Manager) Configure(
	keyConfig map[command.Command]keyinfo.KeyData,
) error {
	copied := make(
		map[command.Command]keyinfo.KeyData,
		len(keyConfig),
	)

	for cmd, binding := range keyConfig {
		copied[cmd] = keyinfo.NewKeyData(
			binding.KeyCode,
			binding.LocaleName,
			append([]string(nil), binding.Modifiers...),
			append([]string(nil), binding.ModifierLocaleNames...),
		)
	}

	m.mu.Lock()
	m.keyConfig = copied
	m.mu.Unlock()

	logger.Debug(
		logModule,
		fmt.Sprintf(
			"macOS global hotkeys configured: %d bindings",
			len(copied),
		),
	)

	return nil
}

func (m *Manager) Enable() error {
	m.mu.Lock()

	if !m.started {
		m.mu.Unlock()
		return fmt.Errorf(
			"macOS hotkey provider is not started",
		)
	}

	m.enabled = true

	m.mu.Unlock()

	logger.Info(
		logModule,
		"macOS global hotkeys enabled",
	)

	return nil
}

func (m *Manager) Disable() error {
	m.mu.Lock()
	m.enabled = false
	m.mu.Unlock()

	logger.Info(
		logModule,
		"macOS global hotkeys disabled",
	)

	return nil
}

func (m *Manager) SetCommandCallback(
	callback func(command.Command),
) {
	m.mu.Lock()
	m.callback = callback
	m.mu.Unlock()
}

func (m *Manager) Close() error {
	m.mu.Lock()

	if !m.started {
		m.mu.Unlock()
		return nil
	}

	done := m.done
	m.enabled = false

	m.mu.Unlock()

	logger.Info(
		logModule,
		"closing macOS global hotkey provider",
	)

	C.hk_stop()

	<-done

	m.mu.Lock()
	m.started = false
	m.mu.Unlock()

	return nil
}

func darwinModifiers(mask uint32) []string {
	modifiers := make([]string, 0, 4)

	if mask&darwinControlMask != 0 {
		modifiers = append(modifiers, "CTRL")
	}

	if mask&darwinAltMask != 0 {
		modifiers = append(modifiers, "ALT")
	}

	if mask&darwinShiftMask != 0 {
		modifiers = append(modifiers, "SHIFT")
	}

	if mask&darwinMetaMask != 0 {
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

	// KeyCode is the authoritative identity for the binding.
	// LocaleName is retained for display/configuration purposes.
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
			result = append(
				result,
				strings.ToUpper(modifier),
			)
		}
	}

	return result
}
