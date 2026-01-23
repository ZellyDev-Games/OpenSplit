package autosplitter

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/logger"
)

const logModule = "autosplitter"
const magic0, magic1, magic2, magic3 = 'O', 'S', 'R', 'C'

type Socket struct {
	dispatcher *dispatcher.Service
	port       uint16
	mu         sync.Mutex
	conn       net.PacketConn
	closeOnce  sync.Once
	closed     chan struct{}
}

func NewSocket(d *dispatcher.Service, port uint16) *Socket {
	return &Socket{
		dispatcher: d,
		port:       port,
		closed:     make(chan struct{}),
	}
}

func (s *Socket) Close() error {
	var err error
	s.closeOnce.Do(func() {
		close(s.closed)

		s.mu.Lock()
		c := s.conn
		s.conn = nil
		s.mu.Unlock()

		if c != nil {
			err = c.Close()
		}
	})
	return err
}

func (s *Socket) Listen() {
	conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		logger.Errorf(logModule, "ListenPacket err: %v", err)
		return
	}

	s.mu.Lock()
	s.conn = conn
	s.mu.Unlock()

	defer func() {
		_ = s.Close()
	}()

	buf := make([]byte, 7)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			// If we're shutting down, ReadFrom will unblock with an error.
			// Return nil to indicate graceful shutdown.
			select {
			case <-s.closed:
				return
			default:
			}

			if errors.Is(err, net.ErrClosed) {
				return
			}

			logger.Errorf(logModule, "read error: %s", err.Error())
			continue
		}

		if n < 7 {
			logger.Warnf(logModule, "short packet: %d bytes", n)
			continue
		}

		packet := buf[:n]
		if packet[0] != magic0 || packet[1] != magic1 || packet[2] != magic2 || packet[3] != magic3 {
			logger.Warnf(logModule, "invalid magic header")
			continue
		}

		version := int(packet[4])
		ackRequested := int(packet[5]) == 1
		command := dispatcher.Command(packet[6])

		if version != 1 {
			logger.Errorf(logModule, "invalid version: %d", version)
			if ackRequested {
				sendAck(conn, addr, 1)
			}
			continue
		}

		_, err = s.dispatcher.Dispatch(command, nil)
		if err != nil {
			if ackRequested {
				sendAck(conn, addr, 2)
			}
			continue
		}

		if ackRequested {
			sendAck(conn, addr, 0)
		}
	}
}

func sendAck(conn net.PacketConn, addr net.Addr, status byte) {
	buf := make([]byte, 0, 7)
	buf = append(buf, 'O', 'S', 'R', 'C')
	buf = append(buf, 1)    // version
	buf = append(buf, 0x80) // AckRecordType
	buf = append(buf, status)
	_, _ = conn.WriteTo(buf, addr)
}
