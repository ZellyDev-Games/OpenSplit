package nwa

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

type ErrorKind int

const (
	InvalidError ErrorKind = iota
	InvalidCommand
	InvalidArgument
	NotAllowed
	ProtocolError
)

type NWAError struct {
	Kind   ErrorKind
	Reason string
}

type AsciiReply interface{}

type AsciiOk struct{}

type AsciiHash map[string]string

type AsciiListHash []map[string]string

type Ok struct{}

type Hash map[string]string

type ListHash []map[string]string

type EmulatorReply interface{}

type Ascii struct {
	Reply AsciiReply
}

type Error struct {
	Err NWAError
}

type Binary struct {
	Data []byte
}

type NWASyncClient struct {
	Connection net.Conn
	Port       uint32
	Addr       net.Addr
}

func Connect(ip string, port uint32) (*NWASyncClient, error) {
	address := fmt.Sprintf("%s:%d", ip, port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("can't resolve address: %w", err)
	}

	conn, err := net.DialTimeout("tcp", tcpAddr.String(), time.Millisecond*1000)
	if err != nil {
		return nil, err
	}

	return &NWASyncClient{
		Connection: conn,
		Port:       port,
		Addr:       tcpAddr,
	}, nil
}

func (c *NWASyncClient) GetReply() (EmulatorReply, error) {
	readStream := bufio.NewReader(c.Connection)
	firstByte, err := readStream.ReadByte()
	if err != nil {
		if err == io.EOF {
			return nil, errors.New("connection aborted")
		}
		return nil, err
	}

	if firstByte == '\n' {
		mapResult := make(map[string]string)
		for {
			line, err := readStream.ReadBytes('\n')
			if err != nil {
				return nil, err
			}
			if len(line) == 0 {
				break
			}
			if line[0] == '\n' && len(mapResult) == 0 {
				return AsciiOk{}, nil
			}
			if line[0] == '\n' {
				break
			}
			colonIndex := bytes.IndexByte(line, ':')
			if colonIndex == -1 {
				return nil, errors.New("malformed line, missing ':'")
			}
			key := strings.TrimSpace(string(line[:colonIndex]))
			value := strings.TrimSpace(string(line[colonIndex+1 : len(line)-1])) // remove trailing \n
			mapResult[key] = value
		}
		if _, ok := mapResult["error"]; ok {
			reason, hasReason := mapResult["reason"]
			errorStr, hasError := mapResult["error"]
			if hasReason && hasError {
				var mkind ErrorKind
				switch errorStr {
				case "protocol_error":
					mkind = ProtocolError
				case "invalid_command":
					mkind = InvalidCommand
				case "invalid_argument":
					mkind = InvalidArgument
				case "not_allowed":
					mkind = NotAllowed
				default:
					mkind = InvalidError
				}
				return NWAError{
					Kind:   mkind,
					Reason: reason,
				}, nil
			} else {
				return NWAError{
					Kind:   InvalidError,
					Reason: "Invalid reason",
				}, nil
			}
		}
		return Hash(mapResult), nil
	}

	if firstByte == 0 {
		// Binary reply
		header := make([]byte, 4)
		n, err := io.ReadFull(readStream, header)
		if err != nil || n != 4 {
			return nil, errors.New("failed to read header")
		}
		size := binary.BigEndian.Uint32(header)
		data := make([]byte, size)
		_, err = io.ReadFull(readStream, data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	return nil, errors.New("invalid reply")
}

func (c *NWASyncClient) ExecuteCommand(cmd string, argString *string) (EmulatorReply, error) {
	var command string
	if argString == nil {
		command = fmt.Sprintf("%s\n", cmd)
	} else {
		command = fmt.Sprintf("%s %s\n", cmd, *argString)
	}

	_, err := io.WriteString(c.Connection, command)
	if err != nil {
		return nil, err
	}

	return c.GetReply()
}

func (c *NWASyncClient) ExecuteRawCommand(cmd string, argString *string) {
	var command string
	if argString == nil {
		command = fmt.Sprintf("%s\n", cmd)
	} else {
		command = fmt.Sprintf("%s %s\n", cmd, *argString)
	}

	// ignoring error as per TODO in Rust code
	_, _ = io.WriteString(c.Connection, command)
}

func (c *NWASyncClient) sendData(data []byte) {
	buf := make([]byte, 5)
	size := len(data)
	buf[0] = 0
	buf[1] = byte((size >> 24) & 0xFF)
	buf[2] = byte((size >> 16) & 0xFF)
	buf[3] = byte((size >> 8) & 0xFF)
	buf[4] = byte(size & 0xFF)
	// TODO: handle the error
	c.Connection.Write(buf)
	// TODO: handle the error
	c.Connection.Write(data)
}

func (c *NWASyncClient) isConnected() bool {
	// net.Conn in Go does not have a Peek method.
	// We can try to set a read deadline and read with a zero-length buffer to check connection.
	// But zero-length read returns immediately, so we try to read 1 byte with deadline.
	buf := make([]byte, 1)
	c.Connection.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
	n, err := c.Connection.Read(buf)
	if err != nil {
		// If timeout or no data, consider connected
		netErr, ok := err.(net.Error)
		if ok && netErr.Timeout() {
			return true
		}
		return false
	}
	if n > 0 {
		// Data was read, connection is alive
		return true
	}
	return false
}

func (c *NWASyncClient) close() {
	// TODO: handle the error
	c.Connection.Close()
}

func (c *NWASyncClient) reconnected() (bool, error) {
	conn, err := net.DialTimeout("tcp", c.Addr.String(), time.Second)
	if err != nil {
		return false, err
	}
	c.Connection = conn
	return true, nil
}
