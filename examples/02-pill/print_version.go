//go:build !windows

package main

import (
	"fmt"
	"log"
	"runtime"

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

	version, err := getTpmVersion(tpm)
	if err != nil {
		log.Fatalf("getTpmVersion() failed: %v", err)
	}
	fmt.Printf("TPM Version: %s\n", string(version))
}
