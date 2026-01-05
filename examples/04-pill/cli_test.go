package main

import (
	"path/filepath"
	"testing"

	"github.com/google/go-tpm/tpm2/transport/simulator"
	"github.com/loicsikidi/tpm-pills/internal/options"
	"github.com/stretchr/testify/require"
)

// TestCreateLoadWorkflow tests the full create/load workflow:
// 1. Create a signing key
// 2. Load the key back into the TPM
func TestCreateLoadWorkflow(t *testing.T) {
	tpm, err := simulator.OpenSimulator()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, tpm.Close())
	})

	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "key.tpm")

	// 1. Create signing key
	createOpts := &options.CreateKeyOpts{
		OutputDir: tempDir,
		KeyType:   options.Signer.String(),
	}
	err = createCommand(tpm, createOpts)
	require.NoError(t, err)
	require.FileExists(t, keyPath)

	// 2. Load the key back into the TPM
	loadOpts := &LoadKeyOpts{
		KeyBlobPath: keyPath,
	}
	err = loadCommand(tpm, loadOpts)
	require.NoError(t, err)
}
