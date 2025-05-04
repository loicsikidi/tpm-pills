//go:build windows

package tpmutil

import (
	"fmt"

	"github.com/google/go-tpm/tpm2/transport"
	"github.com/google/go-tpm/tpm2/transport/simulator"
	"github.com/google/go-tpm/tpm2/transport/windowstpm"
)

type Device string

var (
	TPM_SIMULATOR Device = "simulator"
	WINDOWS       Device = "windows"
	DefaultDevice Device = TPM_SIMULATOR
)

func OpenTPM(device Device) (transport.TPMCloser, error) {
	if device == "" {
		device = DefaultDevice
	}

	switch device {
	case WINDOWS:
		return windowstpm.Open()
	case TPM_SIMULATOR:
		return simulator.OpenSimulator()
	default:
		return nil, fmt.Errorf("unsupported device: %s", device)
	}
}
