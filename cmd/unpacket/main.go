package main

import (
	"encoding/binary"
	"fmt"
	"time"

	itch "github.com/markwinter/go-finproto/itch/5.0"
	"github.com/markwinter/unpacket"
)

type EventCode uint8

type SystemEvent struct {
	Timestamp      time.Duration `unpack:"offset=5,length=6"`
	StockLocate    uint16        `unpack:"offset=1,length=2"`
	TrackingNumber uint16        `unpack:"offset=3,length=2"`
	EventCode      EventCode     `unpack:"offset=11,length=1"`
}

func (e SystemEvent) Bytes() []byte {
	data := make([]byte, 12)

	data[0] = 'S'
	binary.BigEndian.PutUint16(data[1:3], 0)

	// Order of these fields are important. We write timestamp to 3:11 first to let us write a uint64, then overwrite 3:5 with tracking number
	binary.BigEndian.PutUint64(data[3:11], uint64(e.Timestamp.Nanoseconds()))
	binary.BigEndian.PutUint16(data[3:5], e.TrackingNumber)

	data[11] = byte(e.EventCode)

	return data
}

func main() {
	loc, _ := time.LoadLocation("Europe/London")
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	timeSinceMidnight := now.Sub(midnight)

	event := itch.MakeSystemEvent(timeSinceMidnight, 0, itch.EVENT_START_HOURS)
	data := event.Bytes()

	fmt.Printf("%+v\n", data)

	system_event := &SystemEvent{}
	err := unpacket.Unpack(event.Bytes(), binary.BigEndian, system_event)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", system_event)

	outdata, err := unpacket.Pack(binary.BigEndian, system_event)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", outdata)
}
