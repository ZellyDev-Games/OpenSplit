package racetimegg

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/zellydev-games/opensplit/command"
	"github.com/zellydev-games/opensplit/dispatcher"
	"github.com/zellydev-games/opensplit/logger"
)

const logModule = "racetimegg"
const magic0, magic1, magic2, magic3 = 'O', 'S', 'R', 'C'

type Socket struct {
	dispatcher *dispatcher.Service
	port       uint16
	mu         sync.Mutex
	conn       net.PacketConn
	peer       net.Addr
	closeOnce  sync.Once
	closed     chan struct{}
}

// NotifyDone tells the connected RaceTime.gg integration that the active run
// has finished. The integration registers its UDP address with HELLO packets.
func (s *Socket) NotifyDone() {
	s.mu.Lock()
	conn, peer := s.conn, s.peer
	s.mu.Unlock()
	if conn == nil || peer == nil {
		logger.Debug(logModule, "cannot notify RaceTime.gg: no connected peer")
		return
	}
	packet := []byte{magic0, magic1, magic2, magic3, 1, 0, byte(command.DONE)}
	if _, err := conn.WriteTo(packet, peer); err != nil {
		logger.Errorf(logModule, "failed to notify RaceTime.gg that run is done: %v", err)
	}
}

func NewSocket(d *dispatcher.Service, port uint16) *Socket {
	logger.Debugf(logModule, "creating autosplitter socket on UDP port %d", port)
	return &Socket{
		dispatcher: d,
		port:       port,
		closed:     make(chan struct{}),
	}
}

func (s *Socket) Close() error {
	logger.Debug(logModule, "closing autosplitter socket")
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
	logger.Infof(logModule, "listening on UDP port %d", s.port)

	s.mu.Lock()
	s.conn = conn
	s.mu.Unlock()

	defer func() {
		logger.Infof(logModule, "listening on UDP port %d", s.port)
		_ = s.Close()
	}()

	buf := make([]byte, 15)
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

		s.mu.Lock()
		s.peer = addr
		s.mu.Unlock()

		version := int(packet[4])
		ackRequested := int(packet[5]) == 1
		c := command.Command(packet[6])

		logger.Debugf(
			logModule,
			"received command %v ack=%v",
			c,
			ackRequested,
		)

		if version != 1 {
			logger.Errorf(logModule, "invalid version: %d", version)
			if ackRequested {
				sendAck(conn, addr, 1)
			}
			continue
		}

		var payload string

		switch c {
		case command.SET_RUNTIME_OFFSET:
			if n < 15 {
				logger.Warnf(
					logModule,
					"%v missing int64 payload (%d bytes)",
					c,
					n,
				)

				if ackRequested {
					sendAck(conn, addr, 2)
				}
				continue
			}

			offset := int64(binary.LittleEndian.Uint64(packet[7:15]))
			payload = strconv.FormatInt(offset, 10)

			logger.Debugf(
				logModule,
				"decoded payload=%d",
				offset,
			)
		}

		_, err = s.dispatcher.Dispatch(c, &payload)
		if err != nil {
			logger.Errorf(
				logModule,
				"dispatch failed for %v: %v",
				c,
				err,
			)
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
	logger.Debugf(
		logModule,
		"sending ack status=%d",
		status,
	)
	buf := make([]byte, 0, 7)
	buf = append(buf, 'O', 'S', 'R', 'C')
	buf = append(buf, 1)    // version
	buf = append(buf, 0x80) // AckRecordType
	buf = append(buf, status)
	_, _ = conn.WriteTo(buf, addr)
}
