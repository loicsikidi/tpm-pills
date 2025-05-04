//go:build windows

package main

import (
	"fmt"
	"log"

	"github.com/loicsikidi/tpm-pills/internal/tpmutil"
)

func main() {
	tpm, err := tpmutil.OpenTPM(tpmutil.WINDOWS)
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
