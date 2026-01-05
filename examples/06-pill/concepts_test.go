package main

import (
	"bytes"
	"testing"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport/simulator"
	"github.com/loicsikidi/go-tpm-kit/tpmcrypto"
	"github.com/loicsikidi/go-tpm-kit/tpmutil"
	"github.com/stretchr/testify/require"
)

var (
	hmacKeyTemplate = tpm2.TPMTPublic{
		Type: tpm2.TPMAlgKeyedHash,
		ObjectAttributes: tpm2.TPMAObject{
			SignEncrypt:         true,
			FixedTPM:            true,
			FixedParent:         true,
			SensitiveDataOrigin: true,
			UserWithAuth:        true,
		},
	}
	sealTemplate = tpm2.TPMTPublic{
		Type:    tpm2.TPMAlgKeyedHash,
		NameAlg: tpm2.TPMAlgSHA256,
		ObjectAttributes: tpm2.TPMAObject{
			FixedTPM:     true,
			FixedParent:  true,
			UserWithAuth: true,
			NoDA:         true,
		},
	}
)

// TestUnsealCreatePrimary tests the unsealing of data using a primary key.
func TestUnsealCreatePrimary(t *testing.T) {
	thetpm, err := simulator.OpenSimulator()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, thetpm.Close())
	})

	dataToSeal := []byte("secret")

	t.Run("without password", func(t *testing.T) {
		sealHandle, err := tpmutil.CreatePrimary(thetpm, tpmutil.CreatePrimaryConfig{
			SealingData: dataToSeal,
			InPublic:    sealTemplate,
		})
		if err != nil {
			t.Fatalf("could not create primary key: %v", err)
		}
		defer sealHandle.Close()

		unsealRsp, err := tpm2.Unseal{
			ItemHandle: tpmutil.ToAuthHandle(sealHandle),
		}.Execute(thetpm)
		if err != nil {
			t.Fatalf("could not unseal data: %v", err)
		}

		if !bytes.Equal(dataToSeal, unsealRsp.OutData.Buffer) {
			t.Fatalf("unsealed data does not match got %s, expected %s", unsealRsp.OutData.Buffer, dataToSeal)
		}
	})

	t.Run("with password", func(t *testing.T) {
		password := []byte("mypassword")
		sealHandle, err := tpmutil.CreatePrimary(thetpm, tpmutil.CreatePrimaryConfig{
			SealingData: dataToSeal,
			InPublic:    sealTemplate,
			UserAuth:    password,
		})
		if err != nil {
			t.Fatalf("could not create primary key: %v", err)
		}
		defer sealHandle.Close()

		unsealRsp, err := tpm2.Unseal{
			ItemHandle: tpmutil.ToAuthHandle(sealHandle, tpm2.PasswordAuth(password)),
		}.Execute(thetpm)
		if err != nil {
			t.Fatalf("could not unseal data: %v", err)
		}

		if !bytes.Equal(dataToSeal, unsealRsp.OutData.Buffer) {
			t.Fatalf("unsealed data does not match got %s, expected %s", unsealRsp.OutData.Buffer, dataToSeal)
		}
	})
}

// TestSealDataSizeLimits tests the size limits for sealed data based on NameAlg.
// The maximum size for sealed data is limited by MAX_SYM_DATA (128 bytes) in TPM 2.0,
// which is consistent across all hash algorithms (SHA1, SHA256, SHA384, SHA512).
func TestSealDataSizeLimits(t *testing.T) {
	thetpm, err := simulator.OpenSimulator()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, thetpm.Close())
	})

	skrHandle, err := tpmutil.CreatePrimary(thetpm, tpmutil.CreatePrimaryConfig{
		InPublic: tpmutil.ECCSRKTemplate,
	})
	if err != nil {
		t.Fatalf("could not create primary key: %v", err)
	}
	defer skrHandle.Close()

	tests := []struct {
		nameAlg     tpm2.TPMAlgID
		maxSize     int
		description string
	}{
		{tpm2.TPMAlgSHA1, 128, "SHA1 allows up to 128 bytes"},
		{tpm2.TPMAlgSHA256, 128, "SHA256 allows up to 128 bytes"},
		{tpm2.TPMAlgSHA384, 128, "SHA384 allows up to 128 bytes"},
		{tpm2.TPMAlgSHA512, 128, "SHA512 allows up to 128 bytes"},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			template := sealTemplate
			template.Type = tpm2.TPMAlgKeyedHash

			// Test with data at maximum size - should succeed
			dataAtMax := tpmutil.MustGenerateRnd(tc.maxSize)

			keyHandle, err := tpmutil.Create(thetpm, tpmutil.CreateConfig{
				ParentHandle: skrHandle,
				InPublic:     template,
				SealingData:  dataAtMax,
			})
			if err != nil {
				t.Fatalf("failed to seal data at max size (%d bytes) with %v: %v", tc.maxSize, tc.nameAlg, err)
			}
			defer keyHandle.Close()

			unsealRsp, err := tpm2.Unseal{
				ItemHandle: tpmutil.ToAuthHandle(keyHandle),
			}.Execute(thetpm)
			if err != nil {
				t.Fatalf("failed to unseal data at max size: %v", err)
			}

			if !bytes.Equal(dataAtMax, unsealRsp.OutData.Buffer) {
				t.Fatalf("unsealed data does not match for max size")
			}

			// Test with data exceeding maximum size - should fail
			dataOverMax := tpmutil.MustGenerateRnd(tc.maxSize + 1)

			if _, err := tpmutil.Create(thetpm, tpmutil.CreateConfig{
				ParentHandle: skrHandle,
				InPublic:     template,
				SealingData:  dataOverMax,
			}); err == nil {
				t.Fatalf("expected error when sealing data over max size (%d bytes) with %v, but succeeded", tc.maxSize+1, tc.nameAlg)
			}
		})
	}
}

// TestHMAC demonstrates how to create and use HMAC keys with different hash algorithms.
func TestHMAC(t *testing.T) {
	thetpm, err := simulator.OpenSimulator()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, thetpm.Close())
	})

	tests := []struct {
		name     string
		hashAlg  tpm2.TPMIAlgHash
		wantSize int
	}{
		{
			name:     "sha256",
			hashAlg:  tpm2.TPMAlgSHA256,
			wantSize: 32,
		},
		{
			name:     "sha384",
			hashAlg:  tpm2.TPMAlgSHA384,
			wantSize: 48,
		},
		{
			name:     "sha512",
			hashAlg:  tpm2.TPMAlgSHA512,
			wantSize: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := hmacKeyTemplate
			template.NameAlg = tt.hashAlg
			params, err := tpmcrypto.NewHMACParameters(tt.hashAlg)
			if err != nil {
				t.Fatalf("failed to create HMAC parameters: %v", err)
			}
			template.Parameters = *params
			hmacKeyHandle, err := tpmutil.CreatePrimary(thetpm, tpmutil.CreatePrimaryConfig{
				InPublic: template,
			})
			if err != nil {
				t.Fatalf("failed to create primary key: %v", err)
			}
			defer hmacKeyHandle.Close()

			data := []byte("hello world")
			cfg := tpmutil.HmacConfig{
				KeyHandle: hmacKeyHandle,
				Data:      data,
			}
			result, err := tpmutil.Hmac(thetpm, cfg)
			if err != nil {
				t.Fatalf("HMAC failed: %v", err)
			}
			if len(result) != tt.wantSize {
				t.Errorf("HMAC result size = %d, want %d", len(result), tt.wantSize)
			}

			result2, err := tpmutil.Hmac(thetpm, cfg)
			if err != nil {
				t.Fatalf("HMAC failed: %v", err)
			}
			if !bytes.Equal(result, result2) {
				t.Errorf("HMAC results do not match")
			}
		})
	}
}
