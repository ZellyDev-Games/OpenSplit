package main

import (
	"encoding/binary"
	"net"
)

func main() {
	conn, err := net.Dial("udp", "127.0.0.1:6767")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buf := make([]byte, 0, 64)

	// ---- Header ----
	buf = append(buf, 'O', 'S', 'F', 'F') // magic
	buf = append(buf, 1)                  // version
	buf = append(buf, 0)                  // flags (no ack, no crc)

	seq := uint32(1)
	buf = append(buf,
		byte(seq>>24),
		byte(seq>>16),
		byte(seq>>8),
		byte(seq),
	)

	// ---- TLV: SetUnsigned ----
	buf = append(buf, 1) // RecordType = SetUnsigned

	valueLen := uint16(8 + 8) // RecordID (8) + uint64 (8)
	buf = append(buf, byte(valueLen>>8), byte(valueLen))

	// RecordID = "STAGE" padded to 8 bytes
	id := [8]byte{}
	copy(id[:], []byte("STAGE"))
	buf = append(buf, id[:]...)

	// uint64 value = 1
	val := uint64(2)
	tmp := make([]byte, 8)
	binary.BigEndian.PutUint64(tmp, val)
	buf = append(buf, tmp...)

	// ---- Send ----
	_, err = conn.Write(buf)
	if err != nil {
		panic(err)
	}
}
