package netio

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"

	"github.com/google/gopacket/pcap"
)

// IoInterface defines an interface for sending and receiving network frames.
type IoInterface interface {
	SendFrame(frame []byte) (int, error)
	ReceiveFrame(byteSize int, chn chan []byte, errChn chan error)
}

// NewIoInterface creates a new network I/O interface.
// If 'list' is true, it lists the available network interfaces and prompts the user to select one.
// Otherwise, it uses the specified 'interfaceName' to initialize the interface.
func NewIoInterface(list bool, interfaceName string) (IoInterface, error) {
	var (
		nic string
		err error
	)

	if list {
		// List interfaces and allow the user to choose one.
		nic, err = listAndChooseInterface()
		if err != nil {
			return nil, err
		}
	} else {
		// Use the specified interface name.
		if interfaceName == "" {
			return nil, fmt.Errorf("interface name must be provided if list is false")
		}
		nic = interfaceName
	}

	return NewIO(nic), nil
}

// isWindows returns true if the current operating system is Windows.
func isWindows() bool {
	return runtime.GOOS == "windows"
}

// isUnixLike returns true if the current operating system is Unix-like (Linux or macOS).
func isUnixLike() bool {
	return runtime.GOOS == "linux" || runtime.GOOS == "darwin"
}

// listAndChooseInterface lists the available network interfaces on the system and prompts the user to select one.
// It supports both Windows and Unix-like systems.
func listAndChooseInterface() (string, error) {
	var choice int

	if isWindows() {
		// List network interfaces on Windows.
		devices, err := pcap.FindAllDevs()

		if err != nil {
			return "", err
		}

		log.Println("Available network interfaces:")

		for i, dev := range devices {
			log.Printf("%d: %s %s \n", i+1, dev.Name, dev.Description)
		}

		// Ask the user to choose an interface.
		fmt.Print("Enter the number of the interface you want to use: ")

		_, err = fmt.Scanf("%d", &choice)
		if err != nil {
			return "", err
		}

		if choice < 1 || choice > len(devices) {
			return "", fmt.Errorf("invalid choice")
		}

		return devices[choice-1].Name, nil

	} else if isUnixLike() {
		// List network interfaces on Unix-like systems.
		var interfaces []string

		cmd := exec.Command("ip", "-br", "link")

		output, err := cmd.Output()

		if err != nil {
			return "", err
		}

		lines := strings.Split(string(output), "\n")

		for _, line := range lines {

			if len(line) > 0 {
				parts := strings.Fields(line)

				if len(parts) > 0 {
					interfaces = append(interfaces, parts[0])
				}
			}
		}

		log.Println("Available network interfaces:")

		for i, device := range interfaces {
			fmt.Printf("%d: %s\n", i+1, device)
		}

		// Ask the user to choose an interface.
		fmt.Print("Enter the number of the interface you want to use: ")

		_, err = fmt.Scanf("%d", &choice)

		if err != nil {
			return "", err
		}

		if choice < 1 || choice > len(interfaces) {
			return "", fmt.Errorf("invalid choice")
		}

		return interfaces[choice-1], nil
	}

	return "", fmt.Errorf("unsupported operating system")
}

// WindowsIO represents network I/O for Windows systems.
type WindowsIO struct {
	Interface string
}

// UnixLikeIO represents network I/O for Unix-like systems (Linux/macOS).
type UnixLikeIO struct {
	Interface string
}

// NewIO creates a new IoInterface based on the current operating system and selected network interface.
func NewIO(device string) IoInterface {
	if isWindows() {
		return &WindowsIO{
			Interface: device,
		}
	} else if isUnixLike() {
		return &UnixLikeIO{
			Interface: device,
		}
	}
	return nil
}
