//go:build windows

package main

import (
	"fmt"
	"log"

	"github.com/google/go-tpm/tpm2/transport/windowstpm"
)

func main() {
	tpm, err := windowstpm.Open()
	if err != nil {
		log.Fatalf("can't open tpm: %v", err)
	}
	defer tpm.Close()

	version, err := getTpmVersion(tpm)
	if err != nil {
		log.Fatalf("getTpmVersion() failed: %v", err)
	}
	fmt.Printf("TPM Version: %s\n", string(version))
}
