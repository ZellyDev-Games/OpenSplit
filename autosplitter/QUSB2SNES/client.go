package qusb2snes

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/zellydev-games/opensplit/logger"
)

type MessageReaderWriter interface {
	WriteMessage(data []byte) error
	ReadMessage() (p []byte, err error)
	Connect(url.URL) (bool, error)
}

type Command int

const (
	AppVersion Command = iota
	Name
	DeviceList
	Attach
	InfoCommand
	Boot
	Reset
	Menu
	List
	PutFile
	GetFile
	Rename
	Remove
	GetAddress
)

func (c Command) String() string {
	return [...]string{
		"AppVersion",
		"Name",
		"DeviceList",
		"Attach",
		"Info",
		"Boot",
		"Reset",
		"Menu",
		"List",
		"PutFile",
		"GetFile",
		"Rename",
		"Remove",
		"GetAddress",
	}[c]
}

type Space int

const (
	CMD Space = iota
	SNES
)

func (s Space) String() string {
	return [...]string{
		"CMD",
		"SNES",
	}[s]
}

type Info struct {
	Version string
	DevType string
	Game    string
	Flags   []string
}

type USB2SnesQuery struct {
	Opcode   string   `json:"Opcode"`
	Space    string   `json:"Space,omitempty"`
	Flags    []string `json:"Flags,omitempty"`
	Operands []string `json:"Operands"`
}

type USB2SnesResult struct {
	Results []string `json:"Results"`
}

type USB2SnesFileType int

type SyncClient struct {
	messageReaderWriter MessageReaderWriter
	devel               bool
}

func NewSyncClient(messageReaderWriter MessageReaderWriter, devel bool) *SyncClient {
	return &SyncClient{
		devel:               devel,
		messageReaderWriter: messageReaderWriter,
	}
}

func (c *SyncClient) Connect(host string, port uint16) error {
	targetURL := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   "/",
	}

	_, err := c.messageReaderWriter.Connect(targetURL)
	return err
}

func (c *SyncClient) SetName(name string) error {
	return c.sendCommand(Name, CMD, name)
}

func (c *SyncClient) AppVersion() (string, error) {
	err := c.sendCommand(AppVersion, CMD)
	if err != nil {
		return "", err
	}
	reply, err := c.getReply()
	if err != nil {
		return "", err
	}
	if len(reply.Results) == 0 {
		return "", fmt.Errorf("no results in reply")
	}
	return reply.Results[0], nil
}

func (c *SyncClient) ListDevice() ([]string, error) {
	err := c.sendCommand(DeviceList, CMD)
	if err != nil {
		return nil, err
	}
	reply, err := c.getReply()
	if err != nil {
		return nil, err
	}
	return reply.Results, nil
}

func (c *SyncClient) Attach(device string) error {
	return c.sendCommand(Attach, CMD, device)
}

func (c *SyncClient) Info() (*Info, error) {
	err := c.sendCommand(InfoCommand, CMD)
	if err != nil {
		return nil, err
	}
	usbReply, err := c.getReply()
	if err != nil {
		return nil, err
	}
	info := usbReply.Results
	if len(info) < 3 {
		return nil, fmt.Errorf("unexpected reply length")
	}
	var flags []string
	if len(info) > 3 {
		flags = info[3:]
	}
	return &Info{
		Version: info[0],
		DevType: info[1],
		Game:    info[2],
		Flags:   flags,
	}, nil
}

func (c *SyncClient) Reset() error {
	return c.sendCommand(Reset, CMD)
}

func (c *SyncClient) GetAddresses(pairs [][2]int) ([][]byte, error) {
	args := make([]string, 0, len(pairs)*2)
	totalSize := 0
	for _, pair := range pairs {
		address := pair[0]
		size := pair[1]
		args = append(args, strings.ToUpper(fmt.Sprintf("%x", address)))
		args = append(args, fmt.Sprintf("%x", size))
		totalSize += size
	}

	err := c.sendCommand(GetAddress, SNES, args...)
	if err != nil {
		return nil, err
	}

	data := make([]byte, 0, totalSize)
	ret := make([][]byte, 0, len(pairs))

	for {
		msgData, err := c.messageReaderWriter.ReadMessage()
		if err != nil {
			return nil, err
		}

		data = append(data, msgData...)

		if len(data) == totalSize {
			break
		}
	}

	consumed := 0
	for _, pair := range pairs {
		size := pair[1]
		ret = append(ret, data[consumed:consumed+size])
		consumed += size
	}

	return ret, nil
}

func (c *SyncClient) sendCommand(command Command, space Space, args ...string) error {
	if c.devel {
		logger.Debug(
			fmt.Sprintf(
				"Send command : %s %s\n", command.String(), strings.Join(args, " ")))
	}

	query := USB2SnesQuery{
		Opcode:   command.String(),
		Space:    space.String(),
		Flags:    []string{},
		Operands: args,
	}

	jsonData, err := json.Marshal(query)
	if err != nil {
		return err
	}
	if c.devel {
		prettyJSON, err := json.MarshalIndent(query, "", "  ")
		if err == nil {
			fmt.Println(string(prettyJSON))
		}
	}
	err = c.messageReaderWriter.WriteMessage(jsonData)
	return err
}

func (c *SyncClient) getReply() (*USB2SnesResult, error) {
	message, err := c.messageReaderWriter.ReadMessage()
	if err != nil {
		return nil, err
	}
	if c.devel {
		fmt.Println("Reply:")
		fmt.Println(string(message))
	}
	var result USB2SnesResult
	err = json.Unmarshal(message, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
