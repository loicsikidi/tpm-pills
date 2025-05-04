package main

import (
	"bytes"
	"testing"

	"github.com/loicsikidi/tpm-pills/internal/tpmutil"
)

func TestGetTpmVersion(t *testing.T) {
	tpm, err := tpmutil.OpenTPM(tpmutil.TPM_SIMULATOR)
	if err != nil {
		t.Fatalf("can't open tpm: %v", err)
	}
	t.Cleanup(func() {
		if err := tpm.Close(); err != nil {
			t.Fatalf("can't close tpm: %v", err)
		}
	})

	version, err := getTpmVersion(tpm)
	if err != nil {
		t.Fatalf("getTpmVersion() failed: %v", err)
	}

	expectedVersion := []byte{0x32, 0x2E, 0x30, 0x0} // equivalent to '2.0'
	if !bytes.Equal(version, expectedVersion) {
		t.Errorf("expected version %v, got %v", expectedVersion, version)
	}
}
