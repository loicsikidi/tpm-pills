package main

import (
	"crypto/aes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/loicsikidi/go-tpm-kit/tpmtest"
	"github.com/loicsikidi/tpm-pills/internal/options"
	"github.com/stretchr/testify/require"
)

// TestEncryptDecryptWorkflow tests the full encrypt/decrypt workflow:
// 1. Create a key
// 2. Encrypt a message
// 3. Decrypt the blob
func TestEncryptDecryptWorkflow(t *testing.T) {
	tpm := tpmtest.OpenSimulator(t)

	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "key.tpm")
	encryptedPath := filepath.Join(tempDir, "blob.enc")
	message := "secret message"

	// 1. Create key
	createOpts := &options.CreateKeyOpts{
		OutputDir: tempDir,
		KeyType:   options.Decrypt.String(),
	}
	err := createCommand(tpm, createOpts)
	require.NoError(t, err)
	require.FileExists(t, keyPath)

	// 2. Encrypt message
	encryptOpts := &options.SymEncryptOpts{
		KeyBlobPath:    keyPath,
		Message:        message,
		OutputFilePath: encryptedPath,
	}
	err = encryptCommand(tpm, encryptOpts)
	require.NoError(t, err)
	require.FileExists(t, encryptedPath)

	// Verify encrypted blob structure
	data, err := os.ReadFile(encryptedPath)
	require.NoError(t, err)
	var blob encryptedBlob
	err = json.Unmarshal(data, &blob)
	require.NoError(t, err)
	require.NotEmpty(t, blob.Ciphertext)
	require.Equal(t, aes.BlockSize, len(blob.IV))

	// 3. Decrypt blob
	decryptOpts := &options.DecryptOpts{
		KeyBlobPath:   keyPath,
		InputFilePath: encryptedPath,
	}
	decrypted, err := decryptCommand(tpm, decryptOpts)
	require.NoError(t, err)
	require.Equal(t, message, string(decrypted))
}

// TestSealUnsealWorkflow tests the full seal/unseal workflow:
// 1. Seal a message
// 2. Unseal the message
func TestSealUnsealWorkflow(t *testing.T) {
	tpm := tpmtest.OpenSimulator(t)

	tempDir := t.TempDir()
	sealedPath := filepath.Join(tempDir, "sealed_key.tpm")
	message := "sealed secret"

	// 1. Seal message
	sealOpts := &options.SealOpts{
		Message:        message,
		OutputFilePath: sealedPath,
	}
	err := sealCommand(tpm, sealOpts)
	require.NoError(t, err)
	require.FileExists(t, sealedPath)

	// 2. Unseal message
	unsealOpts := &options.UnsealOpts{
		InputFilePath: sealedPath,
	}
	unsealed, err := unsealCommand(tpm, unsealOpts)
	require.NoError(t, err)
	require.Equal(t, message, string(unsealed))
}

// TestHMACWorkflow tests the HMAC computation workflow:
// 1. Compute HMAC
func TestHMACWorkflow(t *testing.T) {
	tpm := tpmtest.OpenSimulator(t)

	data := "data to authenticate"

	// 1. Compute HMAC
	hmacOpts := &options.HMACOpts{
		Data: data,
	}
	result, err := hmacCommand(tpm, hmacOpts)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Verify HMAC is deterministic with same key
	result2, err := hmacCommand(tpm, hmacOpts)
	require.NoError(t, err)
	require.Equal(t, result, result2)
}
