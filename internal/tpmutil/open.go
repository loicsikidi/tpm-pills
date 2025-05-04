//go:build !windows

package tpmutil

import (
	"fmt"
	"runtime"
	"slices"

	swtpm "github.com/foxboron/swtpm_test"
	"github.com/google/go-tpm/tpm2/transport"
	"github.com/google/go-tpm/tpm2/transport/linuxtpm"
	"github.com/google/go-tpm/tpm2/transport/simulator"
)

type Device string

var (
	TPM_SIMULATOR Device = "simulator"
	SWTPM         Device = "swtpm"
	LINUX         Device = "/dev/tpmrm0"
)

// OpenTPM opens a connection to the specified TPM device.
// If no device is specified, it defaults to the appropriate device based on the OS.
func OpenTPM(device Device) (transport.TPMCloser, error) {
	if device == "" {
		device = getDefaultDevice()
	}

	if err := validateDevice(device); err != nil {
		return nil, err
	}

	switch device {
	case LINUX:
		return linuxtpm.Open(string(LINUX))
	case SWTPM:
		return swtpm.OpenSwtpm(SWTPM_STATE)
	case TPM_SIMULATOR:
		return simulator.OpenSimulator()
	default:
		return nil, fmt.Errorf("unsupported device: %s", device)
	}
}

func getDefaultDevice() Device {
	switch runtime.GOOS {
	case "darwin":
		return TPM_SIMULATOR
	case "linux":
		return LINUX
	default:
		return TPM_SIMULATOR
	}
}

func validateDevice(device Device) error {
	switch runtime.GOOS {
	case "darwin":
		if !slices.Contains([]Device{TPM_SIMULATOR, SWTPM}, device) {
			return fmt.Errorf("darwin only supports %s and %s", TPM_SIMULATOR, SWTPM)
		}
	case "linux":
		return nil // Linux supports all devices
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return nil
}
