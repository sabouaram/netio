package main

import (
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/sabouaram/netio"
)

func main() {

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
