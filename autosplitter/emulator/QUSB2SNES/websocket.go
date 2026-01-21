package qusb2snes

import (
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zellydev-games/opensplit/logger"
)

var ErrClosed = errors.New("client is closed")

const RetryWait = time.Second * 3

// WebsocketClient wraps a *websocket.Conn with some state and provides some retry logic
// that we attempt to hide from the caller
type WebsocketClient struct {
	m           sync.Mutex
	url         url.URL
	conn        *websocket.Conn
	connected   bool
	reconnectCh chan struct{}
	doneCh      chan struct{}
	closeOnce   sync.Once
	startOnce   sync.Once
}

// NewWebsocketClient returns an unconnected client with a preset URL
func NewWebsocketClient(url url.URL) *WebsocketClient {
	return &WebsocketClient{
		url:         url,
		reconnectCh: make(chan struct{}, 1),
		doneCh:      make(chan struct{}),
	}
}

// Connected returns the connected state of the WebsocketClient
func (w *WebsocketClient) Connected() bool {
	w.m.Lock()
	defer w.m.Unlock()
	return w.connected
}

func (w *WebsocketClient) WriteMessage(data []byte) error {
	w.m.Lock()
	if w.connected == false || w.conn == nil {
		w.m.Unlock()
		return errors.New("WriteMessage called on disconnected client")
	}
	conn := w.conn
	w.m.Unlock()

	err := conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		logger.Errorf(logModule, "WriteMessage error: %v", err)
		w.signalReconnect()
		return err
	}

	return nil
}

func (w *WebsocketClient) ReadMessage() (p []byte, err error) {
	w.m.Lock()
	if w.connected == false || w.conn == nil {
		w.m.Unlock()
		return []byte{}, errors.New("ReadMessage called on disconnected client")
	}
	conn := w.conn
	w.m.Unlock()

	_, message, err := conn.ReadMessage()
	if err != nil {
		logger.Errorf(logModule, "ReadMessage error: %v", err)
		w.signalReconnect()
	}
	return message, err
}

// Connect attempts to establish a websocket connection to the configured URL
// and manages the state, and retry logic in a goroutine
func (w *WebsocketClient) Connect() {
	w.startOnce.Do(func() {
		go func() {
			for {
				conn, err := w.safeConnect()
				if err != nil || conn == nil {
					// closed branch (i.e. doneCh hs been closed, most likely via Close())
					// there is nothing left to do, this client can never be used again
					if errors.Is(err, ErrClosed) {
						return
					}

					// retry branch (i.e. a transient network error occurred, and we might be able to reconnect
					if conn == nil && err == nil {
						logger.Errorf(
							logModule,
							"safeConnect returned nil connection: retying in %f seconds",
							RetryWait.Seconds())
					} else if err != nil {
						logger.Errorf(
							logModule, "safeConnect error: %v; retying in %f seconds", err, RetryWait.Seconds())
					}

					timer := time.NewTimer(RetryWait)
					select {
					case <-w.doneCh:
						// closed during retry.  Same as Closed branch: Caller should never expect this client
						// to be useful again
						timer.Stop()
						return
					case <-timer.C:
						continue
					}
				}

				// everything worked properly, lock and set state.
				w.m.Lock()
				w.conn = conn
				w.connected = true
				w.m.Unlock()

				// wait for an explicit CLose() or a read/write error that triggers a reconnect attempt
				select {
				case <-w.doneCh:
					w.closeConnection()
					return
				case <-w.reconnectCh:
					logger.Infof(logModule, "Reconnecting to %v", w.url)
					w.closeConnection()
				}
			}
		}()
	})
}

// Close cleanly shuts down this client
// Callers should no longer expect this client to be useful
func (w *WebsocketClient) Close() {
	w.closeOnce.Do(func() {
		close(w.doneCh)
	})
}

// signalReconnect is a helper function to send a non-blocking message to the reconnectCh
func (w *WebsocketClient) signalReconnect() {
	select {
	case w.reconnectCh <- struct{}{}:
	default:
	}
}

// safeConnect checks the closed channel before attempting to connect
func (w *WebsocketClient) safeConnect() (*websocket.Conn, error) {
	select {
	case <-w.doneCh:
		return nil, ErrClosed
	default:
	}
	logger.Infof(logModule, "Connecting to %v", w.url)
	conn, _, err := websocket.DefaultDialer.Dial(w.url.String(), nil)
	return conn, err
}

// closeConnection safely captures the underlying connection, clears the state of the WebsocketClient
// then attempts to silently close the underlying websocket.Conn
func (w *WebsocketClient) closeConnection() {
	w.m.Lock()
	c := w.conn
	w.conn = nil
	w.connected = false
	w.m.Unlock()

	if c != nil {
		_ = c.Close()
	}
}
