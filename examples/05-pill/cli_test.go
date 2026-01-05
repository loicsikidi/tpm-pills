package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-tpm/tpm2/transport/simulator"
	"github.com/loicsikidi/tpm-pills/internal/options"
	"github.com/stretchr/testify/require"
)

// TestEncryptDecryptWorkflow tests the full encrypt/decrypt workflow:
// 1. Create a decrypt key
// 2. Encrypt a message using the public key
// 3. Decrypt the blob using the TPM key
func TestEncryptDecryptWorkflow(t *testing.T) {
	tpm, err := simulator.OpenSimulator()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, tpm.Close())
	})

	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "key.tpm")
	publicKeyPath := filepath.Join(tempDir, "public.pem")
	encryptedPath := filepath.Join(tempDir, "blob.enc")
	message := "Hello TPM Pills!"

	// 1. Create decrypt key
	createOpts := &options.CreateKeyOpts{
		OutputDir: tempDir,
		KeyType:   options.Decrypt.String(),
	}
	err = createCommand(tpm, createOpts)
	require.NoError(t, err)
	require.FileExists(t, keyPath)
	require.FileExists(t, publicKeyPath)

	// 2. Encrypt message using the public key
	encryptOpts := &options.EncryptOpts{
		PublicKeyPath:  publicKeyPath,
		Message:        message,
		OutputFilePath: encryptedPath,
	}
	err = encryptCommand(encryptOpts)
	require.NoError(t, err)
	require.FileExists(t, encryptedPath)

	// 3. Decrypt blob using the TPM key
	decryptOpts := &options.AsymDecryptOpts{
		KeyBlobPath:   keyPath,
		InputFilePath: encryptedPath,
	}
	decrypted, err := decryptCommand(tpm, decryptOpts)
	require.NoError(t, err)
	require.Equal(t, message, string(decrypted))
}

// TestSignVerifyWorkflow tests the full sign/verify workflow:
// 1. Create a signer key
// 2. Sign a message
// 3. Verify the signature
func TestSignVerifyWorkflow(t *testing.T) {
	tpm, err := simulator.OpenSimulator()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, tpm.Close())
	})

	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "key.tpm")
	publicKeyPath := filepath.Join(tempDir, "public.pem")
	signaturePath := filepath.Join(tempDir, "message.sig")
	message := "Hello TPM Pills!"

	// 1. Create signer key
	createOpts := &options.CreateKeyOpts{
		OutputDir: tempDir,
		KeyType:   options.Signer.String(),
	}
	err = createCommand(tpm, createOpts)
	require.NoError(t, err)
	require.FileExists(t, keyPath)
	require.FileExists(t, publicKeyPath)

	// 2. Sign message
	signOpts := &options.SignOpts{
		KeyBlobPath:    keyPath,
		Message:        message,
		OutputFilePath: signaturePath,
	}
	err = signCommand(tpm, signOpts)
	require.NoError(t, err)
	require.FileExists(t, signaturePath)

	// Verify signature file is not empty
	sigData, err := os.ReadFile(signaturePath)
	require.NoError(t, err)
	require.NotEmpty(t, sigData)

	// 3. Verify signature
	verifyOpts := &options.VerifyOpts{
		PublicKeyPath: publicKeyPath,
		Message:       message,
		SignaturePath: signaturePath,
	}
	err = verifyCommand(verifyOpts)
	require.NoError(t, err)
}

// TestRestrictedSignerWorkflow tests signing with a restricted key:
// 1. Create a restricted signer key
// 2. Sign a message (uses TPM2_Hash internally)
// 3. Verify the signature
func TestRestrictedSignerWorkflow(t *testing.T) {
	tpm, err := simulator.OpenSimulator()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, tpm.Close())
	})

	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "key.tpm")
	publicKeyPath := filepath.Join(tempDir, "public.pem")
	signaturePath := filepath.Join(tempDir, "message.sig")
	message := "Hello TPM Pills!"

	// 1. Create restricted signer key
	createOpts := &options.CreateKeyOpts{
		OutputDir: tempDir,
		KeyType:   options.RestrictedSigner.String(),
	}
	err = createCommand(tpm, createOpts)
	require.NoError(t, err)
	require.FileExists(t, keyPath)
	require.FileExists(t, publicKeyPath)

	// 2. Sign message with restricted key
	signOpts := &options.SignOpts{
		KeyBlobPath:    keyPath,
		Message:        message,
		OutputFilePath: signaturePath,
	}
	err = signCommand(tpm, signOpts)
	require.NoError(t, err)
	require.FileExists(t, signaturePath)

	// 3. Verify signature
	verifyOpts := &options.VerifyOpts{
		PublicKeyPath: publicKeyPath,
		Message:       message,
		SignaturePath: signaturePath,
	}
	err = verifyCommand(verifyOpts)
	require.NoError(t, err)
}
