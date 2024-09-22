//go:build !windows
// +build !windows

package netio

import (
	"net"
	"syscall"
	"unsafe"
)

// SendFrame sends a raw Ethernet frame using a Unix-like system's network interface.
// It sets the network interface to promiscuous mode for capturing frames.
func (u *UnixLikeIO) SendFrame(frame []byte) (int, error) {
	var (
		fd           int
		n            int
		err          error
		device       *net.Interface
		hardwareAddr [8]byte
	)

	// Create a raw socket for sending Ethernet frames.
	fd, err = syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, 0x0300)
	if err != nil {
		return 0, err
	}

	defer func(fd int) {
		err = syscall.Close(fd)
		if err != nil {
			panic(err)
		}
	}(fd)

	// Get the network interface by name.
	device, err = net.InterfaceByName(u.Interface)
	if err != nil {
		return 0, err
	}

	// Prepare hardware address and socket address.
	copy(hardwareAddr[0:7], device.HardwareAddr[0:7])

	addr := syscall.SockaddrLinklayer{
		Protocol: 0x0300, 
		Ifindex:  device.Index,
		Halen:    uint8(len(device.HardwareAddr)),
		Addr:     hardwareAddr,
	}

	// Bind the socket to the network interface.
	err = syscall.Bind(fd, &addr)
	if err != nil {
		return 0, err
	}

	// Enable promiscuous mode on the interface.
	err = setLsfPromisc(u.Interface, true)
	if err != nil {
		return 0, err
	}

	// Send the raw Ethernet frame.
	n, err = syscall.Write(fd, frame)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// ReceiveFrame listens for incoming Ethernet frames on a Unix-like system's network interface
// and sends the frames through a provided channel. Errors are passed through an error channel.
func (u *UnixLikeIO) ReceiveFrame(byteSize int, chn chan []byte, errChn chan error) {

	var (
		fd  int
		n   int
		err error
	)

	go func() {

		defer close(chn)
		defer close(errChn)

		// Create a raw socket for receiving Ethernet frames.
		fd, err = syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, 0x0300) // ETH_P_ALL
		if err != nil {
			errChn <- err
			return
		}

		defer func(fd int) {
			err = syscall.Close(fd)
			if err != nil {
				panic(err)
			}
		}(fd)

		// Bind the socket to the network interface.
		err = syscall.BindToDevice(fd, u.Interface)
		if err != nil {
			errChn <- err
			return
		}

		// Create a buffer for receiving frames.
		buffer := make([]byte, byteSize)

		// Continuously receive frames and send them through the channel.
		for {
			n, _, err = syscall.Recvfrom(fd, buffer, 0)
			if err != nil {
				errChn <- err
				return
			}

			if n > 0 {
				// Send the raw frame bytes through the channel.
				chn <- buffer[:n]
			}
		}
	}()
}

// setLsfPromisc enables or disables promiscuous mode on the specified network interface.
func setLsfPromisc(name string, m bool) error {

	var (
		closeErr error
		ifl      struct {
			name  [syscall.IFNAMSIZ]byte
			flags uint16
		}
	)

	s, e := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM|syscall.SOCK_CLOEXEC, 0)
	if e != nil {
		return e
	}

	defer func(fd int) {
		if err := syscall.Close(fd); err != nil {
			closeErr = err
		}
	}(s)

	copy(ifl.name[:], name)

	_, _, ep := syscall.Syscall(syscall.SYS_IOCTL, uintptr(s), syscall.SIOCGIFFLAGS, uintptr(unsafe.Pointer(&ifl)))
	if ep != 0 {
		return ep
	}

	if m {
		ifl.flags |= uint16(syscall.IFF_PROMISC)

	} else {
		ifl.flags &^= uint16(syscall.IFF_PROMISC)
	}
	_, _, ep = syscall.Syscall(syscall.SYS_IOCTL, uintptr(s), syscall.SIOCSIFFLAGS, uintptr(unsafe.Pointer(&ifl)))

	if ep != 0 {
		return ep
	}

	if closeErr != nil {
		// Handle the syscall close error
		return closeErr
	}

	return nil
}
