//go:build !windows
// +build !windows

package netio

// SendFrame sends a raw Ethernet frame using a specified Windows network interface with pcap.
func (u *WindowsIO) SendFrame(frame []byte) (int, error) {
	return 0, nil
}

// ReceiveFrame captures raw Ethernet frames on a Windows network interface and sends them through a channel.
// Errors are passed through an error channel.
func (u *WindowsIO) ReceiveFrame(byteSize int, chn chan []byte, errChn chan error) {
	return
}
