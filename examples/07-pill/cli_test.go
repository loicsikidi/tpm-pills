package main

import (
	"path/filepath"
	"testing"

	"github.com/google/go-tpm/tpm2/transport/simulator"
	"github.com/loicsikidi/tpm-pills/internal/options"
	"github.com/stretchr/testify/require"
)

const testHandle = "0x81000010"

// TestPersistReadUnpersistWorkflow tests the full persist/read/unpersist workflow:
// 1. Persist a key at a given handle
// 2. Read the persisted key and verify it matches the saved public key
// 3. Unpersist the key
// 4. Verify the handle is no longer available
func TestPersistReadUnpersistWorkflow(t *testing.T) {
	tpm, err := simulator.OpenSimulator()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, tpm.Close())
	})

	tempDir := t.TempDir()
	pubkeyPath := filepath.Join(tempDir, "public.pem")

	// 1. Persist a key
	persistOpts := &options.PersistOpts{Handle: testHandle, OutputDir: tempDir}
	err = persistCommand(tpm, persistOpts)
	require.NoError(t, err, "failed to persist key")

	// 2. Read and verify the persisted key matches
	readOpts := &options.ReadPersistedOpts{Handle: testHandle, PublicKeyPath: pubkeyPath}
	err = readCommand(tpm, readOpts)
	require.NoError(t, err, "persisted key should match the saved public key")

	// 3. Unpersist the key
	unpersistOpts := &options.UnpersistOpts{Handle: testHandle}
	err = unpersistCommand(tpm, unpersistOpts)
	require.NoError(t, err, "failed to unpersist key")

	// 4. Verify the handle is no longer available
	err = readCommand(tpm, readOpts)
	require.Error(t, err, "reading an unpersisted handle should fail")
}

// TestReadMismatchingKey verifies that the read command fails when the public key
// does not match the persisted key.
func TestReadMismatchingKey(t *testing.T) {
	tpm, err := simulator.OpenSimulator()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, tpm.Close())
	})

	unpersistOpts := &options.UnpersistOpts{Handle: testHandle}

	// 1. Persist a first key and save its public key
	firstDir := t.TempDir()
	firstPersistOpts := &options.PersistOpts{Handle: testHandle, OutputDir: firstDir}
	err = persistCommand(tpm, firstPersistOpts)
	require.NoError(t, err, "failed to persist first key")

	savedPubkeyPath := filepath.Join(firstDir, "public.pem")
	require.NoError(t, err)

	// 2. Unpersist the first key and persist a second (different) key at the same handle
	err = unpersistCommand(tpm, unpersistOpts)
	require.NoError(t, err, "failed to unpersist first key")

	secondDir := t.TempDir()
	secondPersistOpts := &options.PersistOpts{Handle: testHandle, OutputDir: secondDir}
	err = persistCommand(tpm, secondPersistOpts)
	require.NoError(t, err, "failed to persist second key")

	// 3. Read should fail because the file has the first key's pubkey, not the second's
	readOpts := &options.ReadPersistedOpts{Handle: testHandle, PublicKeyPath: savedPubkeyPath}
	err = readCommand(tpm, readOpts)
	require.Error(t, err, "read should fail when public keys do not match")
	require.Contains(t, err.Error(), "do not match")

	// Cleanup
	err = unpersistCommand(tpm, unpersistOpts)
	require.NoError(t, err)
}

// TestReadMissingPubkeyFlag verifies that the read command fails when --pubkey is not provided.
func TestReadMissingPubkeyFlag(t *testing.T) {
	tpm, err := simulator.OpenSimulator()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, tpm.Close())
	})

	readOpts := &options.ReadPersistedOpts{Handle: testHandle}
	err = readCommand(tpm, readOpts)
	require.Error(t, err, "read should fail without --pubkey")
	require.Contains(t, err.Error(), "invalid input: PublicKeyPath is required")
}
