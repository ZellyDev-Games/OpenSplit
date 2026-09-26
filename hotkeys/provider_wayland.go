//go:build linux && wayland

package hotkeys

import (
	"context"

	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/keyinfo"
	"github.com/zellydev-games/opensplit/logger"
)

func SetupHotkeys() *LinuxManager {
	return NewLinuxManager()
}

// LinuxManager implements HotkeyProvider for Linux/Wayland.
//
// Runtime global shortcuts are provided through the XDG GlobalShortcuts
// portal. Hotkey recording is handled separately by the focused Wails
// frontend because the GlobalShortcuts portal does not provide a raw
// keyboard hook.
type LinuxManager struct {
	portal *PortalManager

	keyPressedCallback func(keyinfo.KeyData)
	commandCallback    func(command.Command)

	currentKeyConfig map[command.Command]keyinfo.KeyData
}

func NewLinuxManager() *LinuxManager {
	return &LinuxManager{
		portal: NewPortalManager(),
	}
}

func (m *LinuxManager) StartHook(callback func(data keyinfo.KeyData)) error {
	m.keyPressedCallback = callback
	return nil
}

func (m *LinuxManager) Unhook() error {
	m.keyPressedCallback = nil
	return nil
}

func (m *LinuxManager) SetCommandCallback(
	callback func(command.Command),
) {
	m.commandCallback = callback
	m.portal.SetCommandCallback(callback)
}

func (m *LinuxManager) Start(ctx context.Context) error {
	return m.portal.Start(ctx)
}

func (m *LinuxManager) Configure(
	keyConfig map[command.Command]keyinfo.KeyData,
) error {
	m.currentKeyConfig = make(
		map[command.Command]keyinfo.KeyData,
		len(keyConfig),
	)

	for cmd, keyData := range keyConfig {
		m.currentKeyConfig[cmd] = keyData
	}

	return m.portal.Configure(m.currentKeyConfig)
}

func (m *LinuxManager) Close() error {
	return m.portal.Close()
}

func (m *LinuxManager) Enable() error {
	logger.Debugf("hotkeys", "enabling global hotkeys")

	if err := m.portal.Configure(m.currentKeyConfig); err != nil {
		logger.Errorf("hotkeys", "enable global hotkeys failed: %v", err)
		return err
	}

	logger.Infof("hotkeys", "global hotkeys enabled")
	return nil
}

func (m *LinuxManager) Disable() error {
	return m.portal.Disable()
}
