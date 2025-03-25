//go:build !windows

package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"runtime"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport"
	"github.com/google/go-tpm/tpm2/transport/linuxtpm"
	"github.com/google/go-tpm/tpm2/transport/simulator"
)

func main() {
	var (
		tpm    transport.TPMCloser
		errTpm error
	)
	switch runtime.GOOS {
	case "linux":
		tpm, errTpm = linuxtpm.Open("/dev/tpmrm0")
	case "darwin":
		tpm, errTpm = simulator.OpenSimulator()
	default:
		log.Fatalf("unsupported platform: %s", runtime.GOOS)
	}
	if errTpm != nil {
		log.Fatalf("can't open tpm: %v", errTpm)
	}
	defer tpm.Close()

	getCmd := tpm2.GetCapability{
		Capability:    tpm2.TPMCapTPMProperties,
		Property:      uint32(tpm2.TPMPTFamilyIndicator),
		PropertyCount: 1,
	}
	getRsp, err := getCmd.Execute(tpm)
	if err != nil {
		log.Fatalf("cmd.GetCapability() failed: %v", err)
	}

	props, err := getRsp.CapabilityData.Data.TPMProperties()
	if err != nil {
		log.Fatalf("tpm2.TPMUCapabilities.TPMProperties() failed: %v", err)
	}

	// value is stored on 4-octet
	version := make([]byte, 4)
	binary.BigEndian.PutUint32(version, props.TPMProperty[0].Value)
	fmt.Printf("TPM Version: %s", string(version))
}
