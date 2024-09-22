package netio

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/sabouaram/cobra_ui"
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
	var (
		interfaces []string
		choice     string
		cmd        *exec.Cmd
	)

	if isWindows() {
		// List network interfaces on Windows.
		devices, err := pcap.FindAllDevs()

		if err != nil {
			return "", err
		}

		for _, dev := range devices {
			interfaces = append(interfaces, fmt.Sprintf("%s (%s)", dev.Name, dev.Description))
		}

	} else if isUnixLike() {

		if runtime.GOOS == "linux" {
			cmd = exec.Command("ip", "-br", "link")
		} else if runtime.GOOS == "darwin" {
			cmd = exec.Command("ifconfig")
		}

		if cmd == nil {
			return "", errors.New("unsupported OS")
		}

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
	}

	if len(interfaces) == 0 {
		return "", fmt.Errorf("no network interfaces found")
	}

	// Use cobra_ui for the interactive interface selection.

	ui := cobra_ui.New()
	ui.SetQuestions([]cobra_ui.Question{
		{
			CursorStr: "==>",
			Color:     color.FgCyan,
			Text:      "Choose a network interface:",
			Options:   interfaces,
			Handler: func(input string) error {
				choice = input
				return nil
			},
		},
	})

	ui.RunInteractiveUI()

	return strings.Split(choice, " ")[0], nil
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
