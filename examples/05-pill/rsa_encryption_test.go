package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"flag"
	"hash"
	"os"
	"testing"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport/simulator"
	"github.com/loicsikidi/tpm-pills/internal/keyutil"
	"github.com/loicsikidi/tpm-pills/internal/tpmutil"
	"github.com/stretchr/testify/require"
)

var tag = flag.String("tag", "", "Tag to run specific tests")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestRSAOAEPDecryption(t *testing.T) {
	tpm, err := simulator.OpenSimulator()
	if err != nil {
		t.Fatalf("could not connect to TPM simulator: %v", err)
	}
	t.Cleanup(func() {
		if err := tpm.Close(); err != nil {
			t.Errorf("%v", err)
		}
	})

	createPrimaryCmd := tpm2.CreatePrimary{
		PrimaryHandle: tpm2.TPMRHOwner,
		InPublic:      tpm2.New2B(tpm2.RSASRKTemplate),
	}
	createPrimaryRsp, err := createPrimaryCmd.Execute(tpm)
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Cleanup(func() {
		flushContextCmd := tpm2.FlushContext{FlushHandle: createPrimaryRsp.ObjectHandle}
		if _, err := flushContextCmd.Execute(tpm); err != nil {
			t.Errorf("%v", err)
		}
	})

	createCmd := tpm2.Create{
		ParentHandle: tpm2.NamedHandle{
			Handle: createPrimaryRsp.ObjectHandle,
			Name:   createPrimaryRsp.Name,
		},
		InPublic: tpmutil.RSA2048EncryptTemplate,
	}
	createRsp, err := createCmd.Execute(tpm)
	if err != nil {
		t.Fatalf("%v", err)
	}

	rsaPub, err := keyutil.PublicKey(&createRsp.OutPublic)
	if err != nil {
		t.Fatalf("%v", err)
	}

	message := []byte("secret")

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub.(*rsa.PublicKey), message, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	loadCmd := tpm2.Load{
		ParentHandle: tpm2.NamedHandle{
			Handle: createPrimaryRsp.ObjectHandle,
			Name:   createPrimaryRsp.Name,
		},
		InPrivate: createRsp.OutPrivate,
		InPublic:  createRsp.OutPublic,
	}
	loadRsp, err := loadCmd.Execute(tpm)
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Cleanup(func() {
		flushContextCmd := tpm2.FlushContext{FlushHandle: loadRsp.ObjectHandle}
		if _, err := flushContextCmd.Execute(tpm); err != nil {
			t.Errorf("%v", err)
		}
	})

	decryptCmd := tpm2.RSADecrypt{
		KeyHandle: tpm2.NamedHandle{
			Handle: loadRsp.ObjectHandle,
			Name:   loadRsp.Name,
		},
		CipherText: tpm2.TPM2BPublicKeyRSA{Buffer: ciphertext},
		InScheme: tpm2.TPMTRSADecrypt{
			Scheme: tpm2.TPMAlgOAEP,
			Details: tpm2.NewTPMUAsymScheme(
				tpm2.TPMAlgOAEP,
				&tpm2.TPMSEncSchemeOAEP{
					HashAlg: tpm2.TPMAlgSHA256,
				},
			),
		},
	}
	decryptRsp, err := decryptCmd.Execute(tpm)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !bytes.Equal(message, decryptRsp.Message.Buffer) {
		t.Errorf("want %x got %x", message, decryptRsp.Message.Buffer)
	}
}

func TestRSAOAEPDecryptionFailure(t *testing.T) {
	tpm, err := simulator.OpenSimulator()
	if err != nil {
		t.Fatalf("could not connect to TPM simulator: %v", err)
	}
	t.Cleanup(func() {
		if err := tpm.Close(); err != nil {
			t.Errorf("%v", err)
		}
	})

	createPrimaryCmd := tpm2.CreatePrimary{
		PrimaryHandle: tpm2.TPMRHOwner,
		InPublic:      tpm2.New2B(tpm2.RSASRKTemplate),
	}
	createPrimaryRsp, err := createPrimaryCmd.Execute(tpm)
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Cleanup(func() {
		flushContextCmd := tpm2.FlushContext{FlushHandle: createPrimaryRsp.ObjectHandle}
		if _, err := flushContextCmd.Execute(tpm); err != nil {
			t.Errorf("%v", err)
		}
	})

	createCmd := tpm2.Create{
		ParentHandle: tpm2.NamedHandle{
			Handle: createPrimaryRsp.ObjectHandle,
			Name:   createPrimaryRsp.Name,
		},
		InPublic: tpmutil.RSA2048EncryptTemplate,
	}
	createRsp, err := createCmd.Execute(tpm)
	if err != nil {
		t.Fatalf("%v", err)
	}

	rsaPub, err := keyutil.PublicKey(&createRsp.OutPublic)
	if err != nil {
		t.Fatalf("%v", err)
	}

	message := []byte("secret")

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub.(*rsa.PublicKey), message, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// alter the ciphertext to simulate a failure
	ciphertext[0] = 0x01

	loadCmd := tpm2.Load{
		ParentHandle: tpm2.NamedHandle{
			Handle: createPrimaryRsp.ObjectHandle,
			Name:   createPrimaryRsp.Name,
		},
		InPrivate: createRsp.OutPrivate,
		InPublic:  createRsp.OutPublic,
	}
	loadRsp, err := loadCmd.Execute(tpm)
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Cleanup(func() {
		flushContextCmd := tpm2.FlushContext{FlushHandle: loadRsp.ObjectHandle}
		if _, err := flushContextCmd.Execute(tpm); err != nil {
			t.Errorf("%v", err)
		}
	})

	decryptCmd := tpm2.RSADecrypt{
		KeyHandle: tpm2.NamedHandle{
			Handle: loadRsp.ObjectHandle,
			Name:   loadRsp.Name,
		},
		CipherText: tpm2.TPM2BPublicKeyRSA{Buffer: ciphertext},
		InScheme: tpm2.TPMTRSADecrypt{
			Scheme: tpm2.TPMAlgOAEP,
			Details: tpm2.NewTPMUAsymScheme(
				tpm2.TPMAlgOAEP,
				&tpm2.TPMSEncSchemeOAEP{
					HashAlg: tpm2.TPMAlgSHA256,
				},
			),
		},
	}
	if _, err := decryptCmd.Execute(tpm); err == nil {
		t.Errorf("want error, got nil")
	}
}

func TestOAEPKeySizeLimit(t *testing.T) {
	if *tag != "all-tests" {
		t.Skip("Test skipped: --tag=all-tests not provided")
	}

	type args struct {
		keySize    int
		hashMethod hash.Hash
		maxLen     int
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "RSA_DECRYPT_OAEP_2048_SHA256",
			args: args{
				keySize:    2048,
				hashMethod: sha256.New(),
				maxLen:     190,
			},
		},
		{
			name: "RSA_DECRYPT_OAEP_3072_SHA256",
			args: args{
				keySize:    3072,
				hashMethod: sha256.New(),
				maxLen:     318,
			},
		},
		{
			name: "RSA_DECRYPT_OAEP_4096_SHA256",
			args: args{
				keySize:    4096,
				hashMethod: sha256.New(),
				maxLen:     446,
			},
		},
		{
			name: "RSA_DECRYPT_OAEP_4096_SHA512",
			args: args{
				keySize:    4096,
				hashMethod: sha512.New(),
				maxLen:     382,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privateKey, err := rsa.GenerateKey(rand.Reader, tt.args.keySize)
			require.NoError(t, err)
			publicKey := &privateKey.PublicKey

			message := make([]byte, tt.args.maxLen)
			_, err = rand.Read(message)
			require.NoError(t, err)

			_, err = rsa.EncryptOAEP(tt.args.hashMethod, rand.Reader, publicKey, message, nil)
			require.NoError(t, err)

			// Test with a message larger than the maximum length
			largeMessage := make([]byte, tt.args.maxLen+1)
			_, err = rand.Read(largeMessage)
			require.NoError(t, err)
			_, err = rsa.EncryptOAEP(tt.args.hashMethod, rand.Reader, publicKey, largeMessage, nil)
			require.Error(t, err, "expected error for message larger than max length")
		})
	}
}
