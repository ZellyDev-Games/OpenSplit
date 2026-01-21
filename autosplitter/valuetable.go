package autosplitter

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"net"
	"sync"

	"github.com/zellydev-games/opensplit/logger"
)

const logModule = "autosplitter"
const headerLen = 10
const magic0, magic1, magic2, magic3 = 'O', 'S', 'F', 'F'

type RecordType byte

const (
	SetSigned RecordType = iota
	SetUnsigned
	SetBool
)

type RecordID [8]byte

func (id RecordID) String() string {
	return string(bytes.TrimRight(id[:], "\x00 "))
}

const idLength = len(RecordID{})

type Flags byte

const (
	AckRequested Flags = 1 << iota
	HasCRC32
)

func (f Flags) Has(flag Flags) bool {
	return f&flag != 0
}

const (
	setIntLen  = uint16(idLength + 8)
	setBoolLen = uint16(idLength + 1)
)

type RulesEngine interface {
	RisingSigned(id string, old, new int64)
	FallingSigned(id string, old, new int64)

	RisingUnsigned(id string, old, new uint64)
	FallingUnsigned(id string, old, new uint64)

	Edge(id string, new bool)
}

type ValueChange struct {
	RecordID RecordID
	Type     RecordType
	OldS     int64
	NewS     int64
	OldU     uint64
	NewU     uint64
	OldB     bool
	NewB     bool
}

type ValueTable struct {
	m              sync.RWMutex
	conn           net.PacketConn
	engine         RulesEngine
	signedValues   map[RecordID]int64
	unsignedValues map[RecordID]uint64
	booleanValues  map[RecordID]bool
}

func NewValueTable(engine RulesEngine) *ValueTable {
	return &ValueTable{
		signedValues:   map[RecordID]int64{},
		unsignedValues: map[RecordID]uint64{},
		booleanValues:  map[RecordID]bool{},
		engine:         engine,
	}
}

func (t *ValueTable) Listen() error {
	conn, err := net.ListenPacket("udp", ":6767")
	if err != nil {
		return err
	}
	t.conn = conn

	defer func(conn net.PacketConn) {
		_ = conn.Close()
	}(t.conn)

	buf := make([]byte, 1024)
	for {
		n, addr, err := t.conn.ReadFrom(buf)
		if err != nil {
			logger.Errorf("read error: %s", err.Error())
			continue
		}

		if n < headerLen {
			logger.Warnf(logModule, "short packet: %d bytes", n)
			continue
		}

		packet := buf[:n]
		if packet[0] != magic0 || packet[1] != magic1 || packet[2] != magic2 || packet[3] != magic3 {
			logger.Warnf(logModule, "invalid magic header")
			continue
		}

		version := int(packet[4])
		flags := Flags(packet[5])
		sequence := binary.BigEndian.Uint32(packet[6:10])
		ackRequested := flags.Has(AckRequested)
		crc32Present := flags.Has(HasCRC32)
		end := n

		if version != 1 {
			logger.Errorf(logModule, "invalid version: %d", version)
			if ackRequested {
				sendAck(t.conn, addr, sequence, 1)
			}
			continue
		}

		if crc32Present {
			if end < headerLen+4 {
				logger.Warnf(logModule, "crc flag set but packet too short: %d bytes", end)
				if ackRequested {
					sendAck(t.conn, addr, sequence, 4)
				}
				continue
			}
			end -= 4
			sentCRC := binary.BigEndian.Uint32(packet[end : end+4])
			computedCRC := crc32.ChecksumIEEE(packet[:end])
			if computedCRC != sentCRC {
				logger.Warnf(logModule, "CRC mismatch (got %08x, expected %08x)", computedCRC, sentCRC)
				if ackRequested {
					sendAck(t.conn, addr, sequence, 2)
				}
				continue
			}
		}

		idx := headerLen
		ok := true
		var changes []ValueChange
		t.m.Lock()

	parseTLV:
		for idx < end {
			if idx+3 > end {
				logger.Warnf(logModule, "truncated TLV header")
				ok = false
				break parseTLV
			}

			recordType := RecordType(packet[idx])
			valueLength := binary.BigEndian.Uint16(packet[idx+1 : idx+3])
			idx += 3

			if idx+int(valueLength) > end {
				logger.Warnf(logModule, "truncated TLV value (need %d bytes)", valueLength)
				ok = false
				break parseTLV
			}

			value := packet[idx : idx+int(valueLength)]
			idx += int(valueLength)

			switch recordType {
			case SetSigned:
				if valueLength != setIntLen {
					logger.Warnf(logModule, "invalid SetSigned TLV length %d", valueLength)
					ok = false
					break parseTLV
				}
				var recordID RecordID
				copy(recordID[:], value[0:idLength])
				oldValue := t.signedValues[recordID]
				newValue := int64(binary.BigEndian.Uint64(value[idLength : idLength+8]))
				t.signedValues[recordID] = newValue
				if oldValue != newValue {
					changes = append(changes, ValueChange{
						RecordID: recordID,
						Type:     recordType,
						OldS:     oldValue,
						NewS:     newValue,
					})
				}

			case SetUnsigned:
				if valueLength != setIntLen {
					logger.Warnf(logModule, "invalid SetUnsigned TLV length %d", valueLength)
					ok = false
					break parseTLV
				}
				var recordID RecordID
				copy(recordID[:], value[0:idLength])
				oldValue := t.unsignedValues[recordID]
				newValue := binary.BigEndian.Uint64(value[idLength : idLength+8])
				t.unsignedValues[recordID] = newValue
				if oldValue != newValue {
					changes = append(changes, ValueChange{
						RecordID: recordID,
						Type:     recordType,
						OldU:     oldValue,
						NewU:     newValue,
					})
				}

			case SetBool:
				if valueLength != setBoolLen {
					logger.Warnf(logModule, "invalid SetBool TLV length %d", valueLength)
					ok = false
					break parseTLV
				}
				var recordID RecordID
				copy(recordID[:], value[0:idLength])
				oldValue := t.booleanValues[recordID]
				newValue := value[idLength] != 0
				t.booleanValues[recordID] = newValue
				if oldValue != newValue {
					changes = append(changes, ValueChange{
						RecordID: recordID,
						Type:     recordType,
						OldB:     oldValue,
						NewB:     newValue,
					})
				}

			default:
				continue
			}
		}
		t.m.Unlock()

		if ackRequested {
			status := byte(0)
			if !ok {
				status = 3
			}
			sendAck(t.conn, addr, sequence, status)
		}

		processChanges(t.engine, changes)
	}
}

func (t *ValueTable) snapshot() (map[RecordID]int64, map[RecordID]uint64, map[RecordID]bool) {
	t.m.RLock()
	defer t.m.RUnlock()

	sv := make(map[RecordID]int64, len(t.signedValues))
	uv := make(map[RecordID]uint64, len(t.unsignedValues))
	bv := make(map[RecordID]bool, len(t.booleanValues))

	for k, v := range t.signedValues {
		sv[k] = v
	}

	for k, v := range t.unsignedValues {
		uv[k] = v
	}

	for k, v := range t.booleanValues {
		bv[k] = v
	}

	return sv, uv, bv
}

func sendAck(conn net.PacketConn, addr net.Addr, seq uint32, status byte) {
	buf := make([]byte, 0, 14)

	buf = append(buf, 'O', 'S', 'F', 'F')
	buf = append(buf, 1) // version
	buf = append(buf, 0) // flags
	buf = append(buf,
		byte(seq>>24),
		byte(seq>>16),
		byte(seq>>8),
		byte(seq),
	)

	buf = append(buf, 0x80) // AckRecordType
	buf = append(buf, 0, 1) // Length = 1
	buf = append(buf, status)

	_, _ = conn.WriteTo(buf, addr)
}

func processChanges(engine RulesEngine, changes []ValueChange) {
	for _, change := range changes {
		switch change.Type {
		case SetSigned:
			if change.NewS > change.OldS {
				engine.RisingSigned(change.RecordID.String(), change.OldS, change.NewS)
			} else {
				engine.FallingSigned(change.RecordID.String(), change.OldS, change.NewS)
			}
		case SetUnsigned:
			if change.NewU > change.OldU {
				engine.RisingUnsigned(change.RecordID.String(), change.OldU, change.NewU)
			} else {
				engine.FallingUnsigned(change.RecordID.String(), change.OldU, change.NewU)
			}
		case SetBool:
			engine.Edge(change.RecordID.String(), change.NewB)
		}
	}
}
