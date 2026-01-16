package main

import (
	"flag"
	"fmt"
	"time"

	qusb2snes "github.com/zellydev-games/opensplit/autosplitter/QUSB2SNES"
)

type FactFinder string

const (
	Qusb2snes FactFinder = "qusb2snes"
	NWA       FactFinder = "nwa"
	Retroarch FactFinder = "retroarch"
)

func main() {
	var factFinder string
	flag.StringVar(&factFinder, "ff", string(Qusb2snes), "FactFinder backend")
	flag.Parse()

	if FactFinder(factFinder) != Qusb2snes {
		fmt.Println("FactFinder not supported")
		return
	}

	websocketClient := &qusb2snes.WebsocketClient{}
	c := qusb2snes.NewSyncClient(websocketClient, false)
	err := c.Connect("localhost", 23074)
	if err != nil {
		return
	}
	fmt.Println("FactFinder connected")
	d, err := c.ListDevice()
	err = c.Attach(d[0])
	if err != nil {
		fmt.Println(err)
		return
	}

	info, err := c.Info()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(info)

	for {
		fmt.Println()
		fmt.Println()
		var addresses [][2]int
		address := [2]int{0xF50000, 0x100}
		addresses = append(addresses, address)

		m, err := c.GetAddresses(addresses)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(m[0][0])
		time.Sleep(2 * time.Second)
	}
}
