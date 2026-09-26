//go:build linux && wayland

package hotkeys

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/keyinfo"
	"github.com/zellydev-games/opensplit/logger"
)

const (
	portalBusName = "org.freedesktop.portal.Desktop"
	portalPath    = dbus.ObjectPath("/org/freedesktop/portal/desktop")
	portalIface   = "org.freedesktop.portal.GlobalShortcuts"

	logModule = "hotkeys"
)

type portalShortcut struct {
	ID          string
	Description string
	Trigger     string
	Command     command.Command
}

type portalShortcutBinding struct {
	ID         string
	Properties map[string]dbus.Variant
}

type PortalManager struct {
	mu sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc

	conn    *dbus.Conn
	object  dbus.BusObject
	session dbus.ObjectPath

	signalCh chan *dbus.Signal

	callback func(command.Command)

	shortcuts map[string]portalShortcut

	requestsMu sync.Mutex
	requests   map[dbus.ObjectPath]chan *dbus.Signal
}

func NewPortalManager() *PortalManager {
	return &PortalManager{
		shortcuts: make(map[string]portalShortcut),
		requests:  make(map[dbus.ObjectPath]chan *dbus.Signal),
	}
}

func (p *PortalManager) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.conn != nil {
		return nil
	}

	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("connect to session bus: %w", err)
	}

	p.ctx, p.cancel = context.WithCancel(ctx)
	p.conn = conn
	p.object = conn.Object(
		portalBusName,
		portalPath,
	)

	if err := p.installSignalHandler(); err != nil {
		p.cancel()
		p.cancel = nil
		p.conn = nil
		p.object = nil
		return err
	}

	if err := p.createSession(); err != nil {
		p.cancel()
		p.cancel = nil
		p.conn.Close()
		p.conn = nil
		p.object = nil
		return err
	}

	return nil
}

func (p *PortalManager) createSession() error {
	token := "opensplit"

	options := map[string]dbus.Variant{
		"handle_token":         dbus.MakeVariant(token),
		"session_handle_token": dbus.MakeVariant(token),
	}

	var request dbus.ObjectPath

	err := p.object.Call(
		portalIface+".CreateSession",
		0,
		options,
	).Store(&request)

	if err != nil {
		return fmt.Errorf("CreateSession: %w", err)
	}

	_, results, err := p.waitForResponse(request)
	if err != nil {
		return fmt.Errorf("CreateSession response: %w", err)
	}

	value, ok := results["session_handle"]
	if !ok {
		return fmt.Errorf("CreateSession returned no session_handle")
	}

	session, ok := value.Value().(dbus.ObjectPath)
	if !ok {
		// Some portal implementations historically exposed this as
		// a string, so accept that representation too.
		if s, ok := value.Value().(string); ok {
			session = dbus.ObjectPath(s)
		} else {
			return fmt.Errorf(
				"invalid session_handle type %T",
				value.Value(),
			)
		}
	}

	p.session = session

	logger.Infof(
		logModule,
		"created Wayland global shortcut session %s",
		session,
	)

	return nil
}

func (p *PortalManager) installSignalHandler() error {
	rules := []string{
		"type='signal',interface='org.freedesktop.portal.Request',member='Response'",
		"type='signal',interface='org.freedesktop.portal.GlobalShortcuts',member='Activated'",
		"type='signal',interface='org.freedesktop.portal.GlobalShortcuts',member='Deactivated'",
		"type='signal',interface='org.freedesktop.portal.GlobalShortcuts',member='ShortcutsChanged'",
	}

	for _, rule := range rules {
		if err := p.conn.AddMatchSignal(
			dbus.WithMatchInterface(strings.Split(rule, ",")[1][len("interface='") : len(strings.Split(rule, ",")[1])-1]),
		); err != nil {
			return fmt.Errorf("add D-Bus signal match: %w", err)
		}
	}

	p.signalCh = make(chan *dbus.Signal, 64)
	p.conn.Signal(p.signalCh)

	go p.signalLoop()

	return nil
}

func (p *PortalManager) waitForResponse(
	request dbus.ObjectPath,
) (uint32, map[string]dbus.Variant, error) {
	responseCh := make(chan *dbus.Signal, 1)

	p.requestsMu.Lock()
	p.requests[request] = responseCh
	p.requestsMu.Unlock()

	defer func() {
		p.requestsMu.Lock()
		delete(p.requests, request)
		p.requestsMu.Unlock()
	}()

	select {
	case <-p.ctx.Done():
		return 0, nil, p.ctx.Err()

	case signal := <-responseCh:
		if signal == nil || len(signal.Body) < 2 {
			return 0, nil, fmt.Errorf("invalid portal response")
		}

		response, ok := signal.Body[0].(uint32)
		if !ok {
			return 0, nil, fmt.Errorf(
				"invalid portal response code type %T",
				signal.Body[0],
			)
		}

		results, ok := signal.Body[1].(map[string]dbus.Variant)
		if !ok {
			return response, nil, fmt.Errorf(
				"invalid portal response results type %T",
				signal.Body[1],
			)
		}

		if response != 0 {
			return response, results, fmt.Errorf(
				"portal request failed with response code %d",
				response,
			)
		}

		return response, results, nil
	}
}

func (p *PortalManager) signalLoop() {
	for {
		select {
		case <-p.ctx.Done():
			return

		case signal := <-p.signalCh:
			if signal == nil {
				return
			}

			switch signal.Name {
			case "org.freedesktop.portal.Request.Response":
				p.dispatchRequestResponse(signal)

			case "org.freedesktop.portal.GlobalShortcuts.Activated":
				p.handleActivated(signal)

			case "org.freedesktop.portal.GlobalShortcuts.Deactivated":
				p.handleDeactivated(signal)

			case "org.freedesktop.portal.GlobalShortcuts.ShortcutsChanged":
				p.handleShortcutsChanged(signal)
			}
		}
	}
}

func (p *PortalManager) dispatchRequestResponse(signal *dbus.Signal) {
	if signal.Path == "" {
		return
	}

	p.requestsMu.Lock()
	responseCh := p.requests[signal.Path]
	p.requestsMu.Unlock()

	if responseCh == nil {
		return
	}

	select {
	case responseCh <- signal:
	default:
	}
}

func (p *PortalManager) SetCommandCallback(
	callback func(command.Command),
) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.callback = callback
}

func (p *PortalManager) handleActivated(signal *dbus.Signal) {
	if len(signal.Body) < 3 {
		return
	}

	session, ok := signal.Body[0].(dbus.ObjectPath)
	if !ok {
		return
	}

	shortcutID, ok := signal.Body[1].(string)
	if !ok {
		return
	}

	p.mu.Lock()

	if session != p.session {
		p.mu.Unlock()
		return
	}

	shortcut, ok := p.shortcuts[shortcutID]
	callback := p.callback

	p.mu.Unlock()

	if !ok {
		logger.Warnf(
			logModule,
			"received activation for unknown shortcut %q",
			shortcutID,
		)
		return
	}

	if callback == nil {
		logger.Warnf(
			logModule,
			"received shortcut %q without a callback",
			shortcutID,
		)
		return
	}

	logger.Debugf(
		logModule,
		"global shortcut activated id=%q command=%d",
		shortcutID,
		shortcut.Command,
	)

	callback(shortcut.Command)
}

func (p *PortalManager) handleDeactivated(signal *dbus.Signal) {
	// The application currently has no state associated with
	// deactivation, so there is nothing to do here.
}

func (p *PortalManager) handleShortcutsChanged(signal *dbus.Signal) {
	// KDE may report changes made through the portal UI.
	// Keep our local command mapping; the next configuration
	// operation will reconcile it with the application config.
}

func (p *PortalManager) Configure(
	keyConfig map[command.Command]keyinfo.KeyData,
) error {
	p.mu.Lock()

	logger.Debugf(
		logModule,
		"configuring global shortcuts: session=%s bindings_configured=%t key_count=%d",
		p.session,
		len(keyConfig),
	)

	if p.session == "" {
		p.mu.Unlock()
		return fmt.Errorf("portal session has not been created")
	}

	session := p.session

	p.mu.Unlock()

	shortcuts := make(map[string]portalShortcut)

	commands := []command.Command{
		command.SPLIT,
		command.UNDO,
		command.SKIP,
		command.PAUSE,
		command.RESET,
		command.COMPARISON_LEFT,
		command.COMPARISON_RIGHT,
	}

	for _, cmd := range commands {
		data, ok := keyConfig[cmd]
		if !ok {
			continue
		}

		trigger, err := keyDataToPortalTrigger(data)
		if err != nil {
			logger.Warnf(
				logModule,
				"skipping command %d: %v",
				cmd,
				err,
			)
			continue
		}

		if trigger == "" {
			continue
		}

		id := shortcutID(cmd)

		shortcuts[id] = portalShortcut{
			ID:          id,
			Description: commandDescription(cmd),
			Trigger:     trigger,
			Command:     cmd,
		}
	}

	if len(shortcuts) == 0 {
		return nil
	}

	bindings := make([]portalShortcutBinding, 0, len(shortcuts))

	for _, shortcut := range shortcuts {
		properties := map[string]dbus.Variant{
			"description":       dbus.MakeVariant(shortcut.Description),
			"preferred_trigger": dbus.MakeVariant(shortcut.Trigger),
		}

		bindings = append(bindings, portalShortcutBinding{
			ID:         shortcut.ID,
			Properties: properties,
		})
	}

	if err := p.bindShortcutsForSession(session, bindings); err != nil {
		return err
	}

	p.mu.Lock()
	p.shortcuts = shortcuts
	p.mu.Unlock()

	return nil
}

func (p *PortalManager) bindShortcuts(
	bindings []interface{},
) error {
	call := p.object.Call(
		portalIface+".BindShortcuts",
		0,
		p.session,
		bindings,
		"",
		map[string]dbus.Variant{},
	)

	if call.Err != nil {
		return fmt.Errorf("BindShortcuts: %w", call.Err)
	}

	var request dbus.ObjectPath

	if err := call.Store(&request); err != nil {
		return fmt.Errorf("read BindShortcuts request: %w", err)
	}

	if _, _, err := p.waitForResponse(request); err != nil {
		return err
	}

	return nil
}

func (p *PortalManager) Close() error {
	p.mu.Lock()

	cancel := p.cancel
	conn := p.conn
	p.cancel = nil
	p.conn = nil
	p.object = nil
	p.session = ""
	p.shortcuts = make(map[string]portalShortcut)

	p.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	if conn != nil {
		conn.RemoveSignal(p.signalCh)
		conn.Close()
	}

	return nil
}

func (p *PortalManager) Disable() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.shortcuts = make(map[string]portalShortcut)

	return nil
}

func (p *PortalManager) bindShortcutsForSession(
	session dbus.ObjectPath,
	bindings []portalShortcutBinding,
) error {
	options := map[string]dbus.Variant{
		"handle_token": dbus.MakeVariant("bind"),
	}

	var request dbus.ObjectPath

	err := p.object.Call(
		portalIface+".BindShortcuts",
		0,
		session,
		bindings,
		"", // parent_window
		options,
	).Store(&request)

	if err != nil {
		return fmt.Errorf("BindShortcuts call: %w", err)
	}

	_, _, err = p.waitForResponse(request)
	if err != nil {
		return fmt.Errorf("BindShortcuts response: %w", err)
	}

	return nil
}

func (p *PortalManager) OpenConfiguration() error {
	p.mu.Lock()
	session := p.session
	p.mu.Unlock()

	if session == "" {
		return fmt.Errorf("portal session has not been created")
	}

	return p.configureShortcutsForSession(session)
}

func (p *PortalManager) configureShortcutsForSession(
	session dbus.ObjectPath,
) error {
	options := map[string]dbus.Variant{}

	call := p.object.Call(
		portalIface+".ConfigureShortcuts",
		0,
		session,
		"",
		options,
	)

	if call.Err != nil {
		return fmt.Errorf(
			"ConfigureShortcuts: %w",
			call.Err,
		)
	}

	var request dbus.ObjectPath

	if err := call.Store(&request); err != nil {
		return fmt.Errorf(
			"read ConfigureShortcuts request: %w",
			err,
		)
	}

	_, _, err := p.waitForResponse(request)
	return err
}

func shortcutID(cmd command.Command) string {
	return fmt.Sprintf("command-%d", cmd)
}

func commandDescription(cmd command.Command) string {
	switch cmd {
	case command.SPLIT:
		return "Split"

	case command.UNDO:
		return "Undo Split"

	case command.SKIP:
		return "Skip Split"

	case command.PAUSE:
		return "Pause Run"

	case command.RESET:
		return "Reset Run"

	case command.COMPARISON_LEFT:
		return "Previous Comparison"

	case command.COMPARISON_RIGHT:
		return "Next Comparison"

	default:
		return fmt.Sprintf("OpenSplit command %d", cmd)
	}
}

func keyDataToPortalTrigger(
	key keyinfo.KeyData,
) (string, error) {
	if key.LocaleName == "" {
		return "", fmt.Errorf("empty key")
	}

	parts := make([]string, 0, len(key.ModifierLocaleNames)+1)

	for _, modifier := range key.ModifierLocaleNames {
		switch normalizeModifier(modifier) {
		case "CTRL":
			parts = append(parts, "CTRL")

		case "ALT":
			parts = append(parts, "ALT")

		case "SHIFT":
			parts = append(parts, "SHIFT")

		case "META":
			parts = append(parts, "META")

		default:
			return "", fmt.Errorf(
				"unsupported modifier %q",
				modifier,
			)
		}
	}

	parts = append(parts, normalizeKey(key.LocaleName))

	return strings.Join(parts, "+"), nil
}

func normalizeModifier(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "control":
		return "CTRL"

	case "ctrl":
		return "CTRL"

	case "left control":
		return "CTRL"

	case "right control":
		return "CTRL"

	case "alt":
		return "ALT"

	case "left alt":
		return "ALT"

	case "right alt":
		return "ALT"

	case "shift":
		return "SHIFT"

	case "left shift":
		return "SHIFT"

	case "right shift":
		return "SHIFT"

	case "meta":
		return "META"

	case "super":
		return "META"

	case "windows":
		return "META"

	default:
		return strings.ToUpper(value)
	}
}

func normalizeKey(value string) string {
	value = strings.TrimSpace(value)

	switch strings.ToLower(value) {
	case " ":
		return "SPACE"

	case "escape":
		return "ESC"

	case "return":
		return "ENTER"

	case "enter":
		return "ENTER"

	case "backspace":
		return "BACKSPACE"

	case "tab":
		return "TAB"

	case "delete":
		return "DELETE"

	case "insert":
		return "INSERT"

	case "left":
		return "LEFT"

	case "right":
		return "RIGHT"

	case "up":
		return "UP"

	case "down":
		return "DOWN"
	}

	return strings.ToUpper(value)
}
