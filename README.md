# Unpacket

`Unpacket` provides functionality to unpack and pack byte slices into structs using an `unpacket` struct tag. This was mostly written to simplify
network protocol parsing e.g. https://github.com/markwinter/go-finproto

This currently only supports fixed-size packets/fields. Support for the final field to be variable length is incoming.

## Features

- Parse a byte array into a struct.
- Ability to pack a struct back into a byte array.

## Installation

To use `Unpacket`

```bash
go get github.com/markwinter/unpacket
```

## Usage

Add an `unpack` struct tag to your struct fields with `offset` and `length` tag fields.

The supported fields are variants of `int` `uint` `bool` `time.Duration` `string`

### Unpack a byte slice

```go
package main

import (
	"fmt"
	"log"
    
	"github.com/markwinter/unpacket"
)

type EventCode uint8

type SystemEvent struct {
	Timestamp      time.Duration `unpack:"offset=5,length=6"`
	StockLocate    uint16        `unpack:"offset=1,length=2"`
	TrackingNumber uint16        `unpack:"offset=3,length=2"`
	EventCode      EventCode     `unpack:"offset=11,length=1"`
}

func main() {
    data := []byte{83,0,0,0,0,0,162,215,245,12,16,83}
    
    systemEvent := &SystemEvent{}
    err := unpacket.Unpack(data, binary.BigEndian, systemEvent)
    if err != nil {
        log.Fatalf("failed unpacking data: %w", err)
    }
    
    fmt.Printf("%+v", systemEvent)
    // &{Timestamp:11m39.4078628s StockLocate:0 TrackingNumber:0 EventCode:83}
}
```

## Pack a struct into a byte slice 

```go
package main

import (
	"fmt"
	"log"
    
	"github.com/markwinter/unpacket"
)

type EventCode uint8

type SystemEvent struct {
	Timestamp      time.Duration `unpack:"offset=5,length=6"`
	StockLocate    uint16        `unpack:"offset=1,length=2"`
	TrackingNumber uint16        `unpack:"offset=3,length=2"`
	EventCode      EventCode     `unpack:"offset=11,length=1"`
}

func main() {
    systemEvent := &SystemEvent{
        // ...   
    }

    data, err := unpacket.Pack(binary.BigEndian, systemEvent)
    if err != nil {
        log.Fatalf("failed packing data: %w", err)
    }
    
    fmt.Printf("%+v", data)
    // [0 0 0 0 0 0 162 215 245 12 16 83]
}
```
