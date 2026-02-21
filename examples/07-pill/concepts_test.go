package main

import (
	"bytes"
	"crypto"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/google/go-tpm/tpm2"
	"github.com/loicsikidi/go-tpm-kit/tpmcrypto"
	"github.com/loicsikidi/go-tpm-kit/tpmtest"
	"github.com/loicsikidi/go-tpm-kit/tpmutil"
)

// TestNameGeneration proves that NAME = nameAlg || Hash(nameAlg, publicArea)
func TestNameGeneration(t *testing.T) {
	thetpm := tpmtest.OpenSimulator(t)

	srkHandle, err := tpmutil.CreatePrimary(thetpm, tpmutil.CreatePrimaryConfig{
		InPublic: tpmutil.ECCSRKTemplate,
	})
	if err != nil {
		t.Fatalf("failed to create primary key: %v", err)
	}

	keyHandle, err := tpmutil.Create(thetpm, tpmutil.CreateConfig{
		ParentHandle: srkHandle,
		InPublic:     tpmutil.ECCSRKTemplate,
	})
	if err != nil {
		t.Fatalf("failed to create key: %v", err)
	}
	defer keyHandle.Close()

	publicArea := keyHandle.Public()
	if publicArea == nil {
		t.Fatal("public area is nil")
	}

	// Serialize the public area
	publicBytes := tpm2.Marshal(publicArea)

	// Get the name algorithm
	nameAlg := publicArea.NameAlg

	// Calculate the hash: Hash(nameAlg, publicBytes)
	digest, err := tpmcrypto.GetDigestFromHashAlg(publicBytes, nameAlg)
	if err != nil {
		t.Fatalf("failed to get digest from hash algorithm: %v", err)
	}

	// Construct the expected name: nameAlg (2 bytes) || hash
	var expectedName bytes.Buffer
	if err := binary.Write(&expectedName, binary.BigEndian, nameAlg); err != nil {
		t.Fatalf("failed to write nameAlg: %v", err)
	}
	expectedName.Write(digest)

	actualName := keyHandle.Name().Buffer
	if !bytes.Equal(expectedName.Bytes(), actualName) {
		t.Errorf("name mismatch:\nexpected: %x\ngot: %x", expectedName.Bytes(), actualName)
	}
}

// TestTransientHandle proves that a key created in the TPM is associated with a transient handle.
//
// Notes:
//   - CreatePrimary creates a primary key in the TPM.
//   - Create creates a child key under the primary key.
func TestTransientHandle(t *testing.T) {
	thetpm := tpmtest.OpenSimulator(t)

	srkHandle, err := tpmutil.CreatePrimary(thetpm, tpmutil.CreatePrimaryConfig{
		InPublic: tpmutil.ECCSRKTemplate,
	})
	if err != nil {
		t.Fatalf("failed to create primary key: %v", err)
	}

	if srkHandle.Type() != tpmutil.TransientHandle {
		t.Errorf("expected handle type %v, got %s", tpmutil.TransientHandle, srkHandle.Type())
	}

	if srkHandle.HandleValue() < 0x80000000 || srkHandle.HandleValue() > 0x80FFFFFF {
		t.Errorf("handle value %v out of expected range (0x80000000 - 0x80FFFFFF)", srkHandle.HandleValue())
	}

	keyHandle, err := tpmutil.Create(thetpm, tpmutil.CreateConfig{
		ParentHandle: srkHandle,
		InPublic:     tpmutil.ECCSRKTemplate,
	})
	if err != nil {
		t.Fatalf("failed to create key: %v", err)
	}
	defer keyHandle.Close()

	if keyHandle.Type() != tpmutil.TransientHandle {
		t.Errorf("expected handle type %v, got %s", tpmutil.TransientHandle, keyHandle.Type())
	}

	if keyHandle.HandleValue() < 0x80000000 || keyHandle.HandleValue() > 0x80FFFFFF {
		t.Errorf("handle value %v out of expected range (0x80000000 - 0x80FFFFFF)", keyHandle.HandleValue())
	}
}

// TestNameHandleMismatch proves that a session with a mismatched name and handle fails.
func TestNameHandleMismatch(t *testing.T) {
	thetpm := tpmtest.OpenSimulator(t)

	srkHandle, err := tpmutil.CreatePrimary(thetpm, tpmutil.CreatePrimaryConfig{
		InPublic: tpmutil.ECCSRKTemplate,
	})
	if err != nil {
		t.Fatalf("failed to create primary key: %v", err)
	}

	eccP256Template := tpmutil.MustApplicationKeyTemplate()

	key1Handle, err := tpmutil.Create(thetpm, tpmutil.CreateConfig{
		ParentHandle: srkHandle,
		InPublic:     eccP256Template,
	})
	if err != nil {
		t.Fatalf("failed to create key: %v", err)
	}
	defer key1Handle.Close()

	key2Handle, err := tpmutil.Create(thetpm, tpmutil.CreateConfig{
		ParentHandle: srkHandle,
		InPublic:     eccP256Template,
	})
	if err != nil {
		t.Fatalf("failed to create key: %v", err)
	}
	defer key2Handle.Close()

	if bytes.Equal(key1Handle.Name().Buffer, key2Handle.Name().Buffer) {
		t.Fatalf("expected different names for key1 and key2, got the same")
	}

	if key1Handle.HandleValue() == key2Handle.HandleValue() {
		t.Fatalf("expected different handle values for key1 and key2, got the same")
	}

	pub, err := tpmcrypto.PublicKey(key1Handle.Public())
	if err != nil {
		t.Fatalf("failed to get public key: %v", err)
	}

	sigScheme, err := tpmcrypto.GetSigSchemeFromPublicKey(pub, crypto.SHA256)
	if err != nil {
		t.Fatalf("GetSigSchemeFromPublicKey failed: %v", err)
	}

	digest := sha256.Sum256([]byte("data to sign"))

	t.Run("coherent name and handle", func(t *testing.T) {
		_, err = tpm2.Sign{
			KeyHandle: tpm2.AuthHandle{
				Handle: key1Handle.Handle(),
				Name:   key1Handle.Name(),
				Auth:   tpm2.HMAC(tpm2.TPMAlgSHA256, 16, tpm2.Auth([]byte(""))),
			},
			Digest:     tpm2.TPM2BDigest{Buffer: digest[:]},
			InScheme:   sigScheme,
			Validation: tpmutil.NullTicket,
		}.Execute(thetpm)
		if err != nil {
			t.Errorf("expected sign to succeed, got: %v", err)
		}
	})

	t.Run("mismatched name and handle", func(t *testing.T) {
		_, err = tpm2.Sign{
			KeyHandle: tpm2.AuthHandle{
				Handle: key1Handle.Handle(),
				Name:   key2Handle.Name(),
				Auth:   tpm2.HMAC(tpm2.TPMAlgSHA256, 16, tpm2.Auth([]byte(""))),
			},
			Digest:     tpm2.TPM2BDigest{Buffer: digest[:]},
			InScheme:   sigScheme,
			Validation: tpmutil.NullTicket,
		}.Execute(thetpm)
		if !errors.Is(err, tpm2.TPMRCBadAuth) {
			t.Errorf("expected TPMRCBadAuth error, got: %v", err)
		}
	})
}
