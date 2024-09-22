# netio

`netio` is a high-level Golang API that allows developers to send and receive raw packets over network interfaces.   

This module provides an easy-to-use interface for network programming, making it suitable for applications involving raw packet crafting, sniffing, and analysis.

Important: This module requires administrative privileges to send and receive raw packets. Please ensure you run your application with the necessary permissions to avoid any permission-related issues  

Note: CGO must be enabled to use this module. Make sure to set the CGO_ENABLED environment variable to 1 before building/running your application to ensure proper functionality.
## Features

- **Send and Receive Raw Packets**: Effortlessly send and receive raw Ethernet frames.
- **Interface Selection**: Users can choose which network interface to use or specify it by default.  
- **Cross-Platform Support**: Compatible with Unix-like systems and Windows.

Important: promiscuous mode interface is enabled so all system traffic is intercepted.  


### Prerequisites

- On **Unix-like systems**, you must have **libpcap** installed. You can usually install it via your package manager.


- On **Windows**, ensure **WinPcap**  is installed :
    - [Download WinPcap](https://www.winpcap.org/install/)

## Usage

To use `netio`, check this example:

```go
package main

import (
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/sabouaram/netio"
)

func main() {
	// this will try to list and to let the user choose the interface he wants to use 
	// otherwise if you would specify the interface by default you should do => netio.NewIoInterface(false, "ensp03")
	ioInterface, err := netio.NewIoInterface(true, "")
	if err != nil {
		log.Fatalf("Error initializing network interface: %v", err)
	}

	log.Printf("Selected network interface: %v\n", ioInterface)

	ethLayer := &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x00, 0x1a, 0x2b, 0x3c, 0x4d, 0x5e},
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: 0x0806,
	}

	// Create a new ARP layer
	arpLayer := &layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   net.HardwareAddr{0x00, 0x1A, 0x2B, 0x3C, 0x4D, 0x5E},
		SourceProtAddress: net.IP{192, 168, 1, 1},
		DstHwAddress:      net.HardwareAddr{0x00, 0x1a, 0x2a, 0x3b, 0x4b, 0x5c},
		DstProtAddress:    net.IP{192, 168, 1, 2},
	}

	// Create a buffer to hold the serialized packet
	packetBuffer := gopacket.NewSerializeBuffer()

	// Serialize the layers into the buffer
	if err = gopacket.SerializeLayers(packetBuffer, gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	},
		ethLayer,
		arpLayer,
	); err != nil {
		log.Println("Error serializing packet:", err)
		return
	}

	// Get the serialized byte slice
	packetBytes := packetBuffer.Bytes()

	log.Printf("Serialized packet: %x\n", packetBytes)

	_, err = ioInterface.SendFrame(packetBuffer.Bytes())
	if err != nil {
		log.Fatalf("Error sending the raw packet: %v", err)
	}

	log.Println("raw packet sent")

	frameChannel := make(chan []byte)
	errorChannel := make(chan error)

	go ioInterface.ReceiveFrame(65536, frameChannel, errorChannel)

	for {
		select {
		case frame := <-frameChannel:
			if frame != nil {
				log.Printf("Received frame: %x\n", frame)
			}
		case err = <-errorChannel:
			if err != nil {
				log.Printf("Error receiving frame: %v", err)
				return
			}
		}
	}
}

```

## Support

If you enjoy this module and want to support, consider buying me a coffee! ☕️  
[Buy Me a Coffee](https://buymeacoffee.com/sabouaram)  
