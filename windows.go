//go:build windows
// +build windows

package netio

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// SendFrame sends a raw Ethernet frame using a specified Windows network interface with pcap.
func (w *WindowsIO) SendFrame(frame []byte) (int, error) {
	// Open the network interface for live capture with a snapshot length of 65536 bytes.
	handle, err := pcap.OpenLive(w.Interface, 65536, true, pcap.BlockForever)
	if err != nil {
		return 0, err
	}
	defer handle.Close()

	// Write the raw Ethernet frame to the network interface.
	err = handle.WritePacketData(frame)
	if err != nil {
		return 0, err
	}

	// Return the number of bytes sent.
	return len(frame), nil
}

// ReceiveFrame captures raw Ethernet frames on a Windows network interface and sends them through a channel.
// Errors are passed through an error channel.
func (w *WindowsIO) ReceiveFrame(byteSize int, chn chan []byte, errChn chan error) {
	go func() {

		defer close(chn)
		defer close(errChn)

		// Open the network interface for capturing packets.
		handle, err := pcap.OpenLive(w.Interface, int32(byteSize), true, pcap.BlockForever)
		if err != nil {
			errChn <- err
			return
		}

		defer handle.Close()

		// Create a buffer to store the raw frame bytes.
		buffer := make([]byte, byteSize)

		// Continuously read packets from the network interface.
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

		for packet := range packetSource.Packets() {
			// Extract raw packet bytes.
			frameData := packet.Data()

			if len(frameData) > 0 {
				// Send the raw frame bytes through the channel.
				copy(buffer[:len(frameData)], frameData)
				chn <- buffer[:len(frameData)]
			}
		}
	}()
}
