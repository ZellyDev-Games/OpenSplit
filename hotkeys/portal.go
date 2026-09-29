//go:build linux && wayland

package hotkeys

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/keyinfo"
	"github.com/zellydev-games/opensplit/logger"
)

const (
	portalBusName = "org.freedesktop.portal.Desktop"
	portalPath    = dbus.ObjectPath("/org/freedesktop/portal/desktop")
	portalIface   = "org.freedesktop.portal.GlobalShortcuts"
	registryIface = "org.freedesktop.host.portal.Registry"
	appID         = "com.zellydevgames.opensplit"

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

	bindingsConfigured bool
	enabled            bool

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
	if p.conn != nil {
		p.mu.Unlock()
		return nil
	}
	p.mu.Unlock()

	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("connect to session bus: %w", err)
	}

	portalCtx, cancel := context.WithCancel(ctx)

	p.mu.Lock()
	// Another Start could theoretically have completed while we
	// were connecting to D-Bus.
	if p.conn != nil {
		p.mu.Unlock()
		cancel()
		conn.Close()
		return nil
	}

	p.ctx = portalCtx
	p.cancel = cancel
	p.conn = conn
	p.object = conn.Object(
		portalBusName,
		portalPath,
	)
	p.mu.Unlock()

	// Directly launched host apps may lack the app identity desktop portals
	// use to label and present GlobalShortcuts requests. Register before the
	// first portal call. Older portal versions do not expose this interface.
	err = conn.Object(portalBusName, portalPath).Call(
		registryIface+".Register",
		0,
		appID,
		map[string]dbus.Variant{},
	).Err
	if err != nil {
		dbusErr, ok := err.(dbus.Error)
		if !ok || dbusErr.Name != "org.freedesktop.DBus.Error.UnknownMethod" {
			cancel()
			conn.Close()

			p.mu.Lock()
			p.cancel = nil
			p.conn = nil
			p.object = nil
			p.ctx = nil
			p.mu.Unlock()

			return fmt.Errorf("register portal application ID %q: %w", appID, err)
		}
	}

	if err := p.installSignalHandler(); err != nil {
		cancel()
		conn.Close()

		p.mu.Lock()
		p.cancel = nil
		p.conn = nil
		p.object = nil
		p.ctx = nil
		p.mu.Unlock()

		return err
	}

	if err := p.createSession(); err != nil {
		cancel()
		conn.Close()

		p.mu.Lock()
		p.cancel = nil
		p.conn = nil
		p.object = nil
		p.ctx = nil
		p.mu.Unlock()

		return err
	}

	return nil
}

func (p *PortalManager) createSession() error {
	p.mu.Lock()
	object := p.object
	p.mu.Unlock()

	if object == nil {
		return fmt.Errorf("portal connection has not been created")
	}

	token := portalToken("opensplit")

	options := map[string]dbus.Variant{
		"handle_token":         dbus.MakeVariant(token),
		"session_handle_token": dbus.MakeVariant(token),
	}

	var request dbus.ObjectPath

	err := object.Call(
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
		if s, stringOK := value.Value().(string); stringOK {
			session = dbus.ObjectPath(s)
		} else {
			return fmt.Errorf(
				"invalid session_handle type %T",
				value.Value(),
			)
		}
	}

	p.mu.Lock()
	p.session = session
	p.mu.Unlock()

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

	if !p.enabled || session != p.session {
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
	enabled := p.enabled
	session := p.session
	p.mu.Unlock()

	if !enabled {
		// Configuration is retained by LinuxManager. Do not bind
		// anything while global hotkeys are disabled.
		logger.Debugf(
			logModule,
			"global hotkeys disabled; deferring portal configuration",
		)
		return nil
	}

	if session == "" {
		return fmt.Errorf("portal session has not been created")
	}

	if err := p.recreateSession(); err != nil {
		p.mu.Lock()
		p.enabled = false
		p.mu.Unlock()

		return fmt.Errorf("recreate portal session: %w", err)
	}

	if err := p.configureBindings(keyConfig); err != nil {
		p.mu.Lock()
		p.enabled = false
		p.mu.Unlock()

		return err
	}

	return nil
}

func (p *PortalManager) Close() error {
	p.mu.Lock()

	cancel := p.cancel
	conn := p.conn
	signalCh := p.signalCh

	p.cancel = nil
	p.conn = nil
	p.object = nil
	p.session = ""
	p.shortcuts = make(map[string]portalShortcut)
	p.bindingsConfigured = false
	p.enabled = false
	p.signalCh = nil

	p.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	if conn != nil {
		if signalCh != nil {
			conn.RemoveSignal(signalCh)
		}
		conn.Close()
	}

	return nil
}

func (p *PortalManager) Enable(
	keyConfig map[command.Command]keyinfo.KeyData,
) error {
	p.mu.Lock()
	enabled := p.enabled
	session := p.session
	bound := p.bindingsConfigured
	p.mu.Unlock()

	if enabled {
		return nil
	}

	if session == "" {
		if err := p.createSession(); err != nil {
			return err
		}

		bound = false
	}

	if !bound {
		if err := p.configureBindings(keyConfig); err != nil {
			return err
		}
	}

	p.mu.Lock()
	p.enabled = true
	p.mu.Unlock()

	return nil
}

func (p *PortalManager) Disable() error {
	if err := p.closeSession(); err != nil {
		return err
	}

	p.resetSessionState()

	p.mu.Lock()
	p.enabled = false
	p.mu.Unlock()

	return nil
}

func (p *PortalManager) recreateSession() error {
	if err := p.closeSession(); err != nil {
		return err
	}

	p.resetSessionState()

	return p.createSession()
}

func (p *PortalManager) closeSession() error {
	p.mu.Lock()
	conn := p.conn
	session := p.session
	p.mu.Unlock()

	if conn == nil || session == "" {
		return nil
	}

	sessionObject := conn.Object(
		portalBusName,
		session,
	)

	call := sessionObject.Call(
		"org.freedesktop.portal.Session.Close",
		0,
	)

	if call.Err != nil {
		return fmt.Errorf(
			"close global shortcut session: %w",
			call.Err,
		)
	}

	return nil
}

func (p *PortalManager) bindShortcutsForSession(
	session dbus.ObjectPath,
	bindings []portalShortcutBinding,
) error {
	options := map[string]dbus.Variant{
		"handle_token": dbus.MakeVariant(portalToken("bind")),
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

func (p *PortalManager) configureBindings(
	keyConfig map[command.Command]keyinfo.KeyData,
) error {
	shortcuts, bindings, err := portalShortcutDefinitions(keyConfig)
	if err != nil {
		return err
	}

	p.mu.Lock()
	session := p.session
	p.mu.Unlock()

	if session == "" {
		return fmt.Errorf("portal session has not been created")
	}

	// An empty binding set is valid from the application's perspective.
	// There is no need to call BindShortcuts with an empty list.
	if len(bindings) == 0 {
		p.mu.Lock()
		p.shortcuts = shortcuts
		p.bindingsConfigured = true
		p.mu.Unlock()

		return nil
	}

	if err := p.bindShortcutsForSession(session, bindings); err != nil {
		return err
	}

	p.mu.Lock()
	p.shortcuts = shortcuts
	p.bindingsConfigured = true
	p.mu.Unlock()

	logger.Infof(
		logModule,
		"configured %d global shortcuts for session %s",
		len(shortcuts),
		session,
	)

	return nil
}

func (p *PortalManager) resetSessionState() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.session = ""
	p.shortcuts = make(map[string]portalShortcut)
	p.bindingsConfigured = false
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

func portalToken(prefix string) string {
	return fmt.Sprintf(
		"%s_%d",
		prefix,
		time.Now().UnixNano(),
	)
}

func portalShortcutDefinitions(
	keyConfig map[command.Command]keyinfo.KeyData,
) (
	map[string]portalShortcut,
	[]portalShortcutBinding,
	error,
) {
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

	bindings := make(
		[]portalShortcutBinding,
		0,
		len(shortcuts),
	)

	for _, shortcut := range shortcuts {
		bindings = append(
			bindings,
			portalShortcutBinding{
				ID: shortcut.ID,
				Properties: map[string]dbus.Variant{
					"description": dbus.MakeVariant(
						shortcut.Description,
					),
					"preferred_trigger": dbus.MakeVariant(
						shortcut.Trigger,
					),
				},
			},
		)
	}

	return shortcuts, bindings, nil
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
	if value == " " {
		return "SPACE"
	}

	value = strings.TrimSpace(value)

	switch strings.ToLower(value) {
	case "space":
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
