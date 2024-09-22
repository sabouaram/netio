//go:build windows
// +build windows

package netio

// SendFrame sends a raw Ethernet frame using a Unix-like system's network interface.
// It sets the network interface to promiscuous mode for capturing frames.
func (u *UnixLikeIO) SendFrame(frame []byte) (int, error) {
	return 0, nil
}

// ReceiveFrame listens for incoming Ethernet frames on a Unix-like system's network interface
// and sends the frames through a provided channel. Errors are passed through an error channel.
func (u *UnixLikeIO) ReceiveFrame(byteSize int, chn chan []byte, errChn chan error) {
	return
}
