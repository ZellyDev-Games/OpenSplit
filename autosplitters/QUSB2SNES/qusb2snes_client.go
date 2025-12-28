package qusb2snes

// TODO: handle errors correctly

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
)

type Command int

const (
	AppVersion Command = iota
	Name
	DeviceList
	Attach
	Info
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
	None Space = iota
	SNES
)

func (s Space) String() string {
	return [...]string{
		"None",
		"SNES",
	}[s]
}

type Infos struct {
	Version string
	DevType string
	Game    string
	Flags   []string
}

type USB2SnesQuery struct {
	Opcode   string   `json:"Opcode"`
	Space    string   `json:"Space,omitempty"`
	Flags    []string `json:"Flags"`
	Operands []string `json:"Operands"`
}

type USB2SnesResult struct {
	Results []string `json:"Results"`
}

type USB2SnesFileType int

const (
	File USB2SnesFileType = iota
	Dir
)

type USB2SnesFileInfo struct {
	Name     string
	FileType USB2SnesFileType
}

type SyncClient struct {
	Client *websocket.Conn
	devel  bool
}

func Connect(host string, port uint32) (*SyncClient, error) {
	return connect(host, port, false)
}

func ConnectWithDevel(host string, port uint32) (*SyncClient, error) {
	return connect(host, port, true)
}

func connect(host string, port uint32, devel bool) (*SyncClient, error) {
	numStr := strconv.FormatUint(uint64(port), 10)
	u := url.URL{Scheme: "ws", Host: host + ":" + numStr, Path: "/"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return &SyncClient{
		Client: conn,
		devel:  devel,
	}, nil
}

func (sc *SyncClient) sendCommand(command Command, args []string) error {
	return sc.sendCommandWithSpace(command, None, args)
}

func (sc *SyncClient) sendCommandWithSpace(command Command, space Space, args []string) error {
	if sc.devel {
		fmt.Printf("Send command : %s\n", command.String())
	}
	// var nspace *string
	// if space != nil {
	// s := space.String()
	// nspace = &s
	// }
	var query USB2SnesQuery
	if space == SNES {
		query = USB2SnesQuery{
			Opcode:   command.String(),
			Space:    space.String(),
			Flags:    []string{},
			Operands: args,
		}
	} else {
		query = USB2SnesQuery{
			Opcode: command.String(),
			// Space:    space.String(),
			Flags:    []string{},
			Operands: args,
		}
	}
	jsonData, err := json.Marshal(query)
	if err != nil {
		return err
	}
	if sc.devel {
		prettyJSON, err := json.MarshalIndent(query, "", "  ")
		if err == nil {
			fmt.Println(string(prettyJSON))
		}
	}
	err = sc.Client.WriteMessage(websocket.TextMessage, jsonData)
	return err
}

func (sc *SyncClient) getReply() (*USB2SnesResult, error) {
	_, message, err := sc.Client.ReadMessage()
	if err != nil {
		return nil, err
	}
	if sc.devel {
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

func (sc *SyncClient) SetName(name string) error {
	return sc.sendCommand(Name, []string{name})
}

func (sc *SyncClient) AppVersion() (string, error) {
	err := sc.sendCommand(AppVersion, []string{})
	if err != nil {
		return "", err
	}
	reply, err := sc.getReply()
	if err != nil {
		return "", err
	}
	if len(reply.Results) == 0 {
		return "", fmt.Errorf("no results in reply")
	}
	return reply.Results[0], nil
}

func (sc *SyncClient) ListDevice() ([]string, error) {
	err := sc.sendCommand(DeviceList, []string{})
	if err != nil {
		return nil, err
	}
	reply, err := sc.getReply()
	if err != nil {
		return nil, err
	}
	return reply.Results, nil
}

func (sc *SyncClient) Attach(device string) error {
	return sc.sendCommand(Attach, []string{device})
}

func (sc *SyncClient) Info() (*Infos, error) {
	err := sc.sendCommand(Info, []string{})
	if err != nil {
		return nil, err
	}
	usbreply, err := sc.getReply()
	if err != nil {
		return nil, err
	}
	info := usbreply.Results
	if len(info) < 3 {
		return nil, fmt.Errorf("unexpected reply length")
	}
	flags := []string{}
	if len(info) > 3 {
		flags = info[3:]
	}
	return &Infos{
		Version: info[0],
		DevType: info[1],
		Game:    info[2],
		Flags:   flags,
	}, nil
}

func (sc *SyncClient) Reset() error {
	return sc.sendCommand(Reset, []string{})
}

// func (sc *SyncClient) Menu() error {
// 	return sc.sendCommand(Menu, []string{})
// }

// func (sc *SyncClient) Boot(toboot string) error {
// 	return sc.sendCommand(Boot, []string{toboot})
// }

// func (sc *SyncClient) Ls(path string) ([]USB2SnesFileInfo, error) {
// 	err := sc.sendCommand(List, []string{path})
// 	if err != nil {
// 		return nil, err
// 	}
// 	usbreply, err := sc.getReply()
// 	if err != nil {
// 		return nil, err
// 	}
// 	vecInfo := usbreply.Results
// 	var toret []USB2SnesFileInfo
// 	for i := 0; i < len(vecInfo); i += 2 {
// 		if i+1 >= len(vecInfo) {
// 			break
// 		}
// 		fileType := Dir
// 		if vecInfo[i] == "1" {
// 			fileType = File
// 		}
// 		info := USB2SnesFileInfo{
// 			FileType: fileType,
// 			Name:     vecInfo[i+1],
// 		}
// 		toret = append(toret, info)
// 	}
// 	return toret, nil
// }

// func (sc *SyncClient) SendFile(path string, data []byte) error {
// 	err := sc.sendCommand(PutFile, []string{path, fmt.Sprintf("%x", len(data))})
// 	if err != nil {
// 		return err
// 	}
// 	chunkSize := 1024
// 	for start := 0; start < len(data); start += chunkSize {
// 		stop := start + chunkSize
// 		if stop > len(data) {
// 			stop = len(data)
// 		}
// 		err = sc.Client.WriteMessage(websocket.BinaryMessage, data[start:stop])
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (sc *SyncClient) getFile(path string) ([]byte, error) {
// 	err := sc.sendCommand(GetFile, []string{path})
// 	if err != nil {
// 		return nil, err
// 	}
// 	reply, err := sc.getReply()
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(reply.Results) == 0 {
// 		return nil, errors.New("no results in reply")
// 	}
// 	stringHex := reply.Results[0]
// 	size, err := strconv.ParseUint(stringHex, 16, 0)
// 	if err != nil {
// 		return nil, err
// 	}
// 	data := make([]byte, 0, size)
// 	for {
// 		_, msgData, err := sc.Client.ReadMessage()
// 		if err != nil {
// 			return nil, err
// 		}
// 		// In Rust code, it expects binary message
// 		// Here, msgData is []byte already
// 		data = append(data, msgData...)
// 		if len(data) == int(size) {
// 			break
// 		}
// 	}
// 	return data, nil
// }

// func (sc *SyncClient) removePath(path string) error {
// 	return sc.sendCommand(Remove, []string{path})
// }

// func (sc *SyncClient) getAddress(address uint32, size int) ([]byte, error) {
// 	err := sc.sendCommandWithSpace(GetAddress, SNES, []string{
// 		fmt.Sprintf("%x", address),
// 		fmt.Sprintf("%x", size),
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	data := make([]byte, 0, size)
// 	for {
// 		_, msgData, err := sc.Client.ReadMessage()
// 		if err != nil {
// 			return nil, err
// 		}
// 		data = append(data, msgData...)
// 		if len(data) == size {
// 			break
// 		}
// 	}
// 	return data, nil
// }

func (sc *SyncClient) getAddresses(pairs [][2]int) ([][]byte, error) {
	args := make([]string, 0, len(pairs)*2)
	totalSize := 0
	for _, pair := range pairs {
		address := pair[0]
		size := pair[1]
		args = append(args, fmt.Sprintf("%x", address))
		args = append(args, fmt.Sprintf("%x", size))
		totalSize += size
	}

	err := sc.sendCommandWithSpace(GetAddress, SNES, args)
	if err != nil {
		return nil, err
	}

	data := make([]byte, 0, totalSize)
	ret := make([][]byte, 0, len(pairs))

	for {
		_, msgData, err := sc.Client.ReadMessage()
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
