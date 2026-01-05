package main

import (
	"bytes"
	"log"
	"testing"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport/simulator"
	"github.com/loicsikidi/tpm-pills/internal/tpmutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReproductability demonstrates that the primary key creation is reproducible
func TestReproductability(t *testing.T) {
	tpm, err := simulator.OpenSimulator()
	if err != nil {
		log.Fatalf("can't open tpm simulator: %v", err)
	}
	defer tpm.Close()

	firstPrimary, closer := setupCreatePrimary(t, tpm, tpm2.New2B(tpmutil.ECCSignerTemplate))
	defer closer()

	secondPrimary, secondCloser := setupCreatePrimary(t, tpm, tpm2.New2B(tpmutil.ECCSignerTemplate))
	defer secondCloser()

	firstEccPub := mustPublicKey(t, firstPrimary)
	secondEccPub := mustPublicKey(t, secondPrimary)

	require.Equal(t, firstEccPub, secondEccPub, "Public keys doesn't match")
	require.True(t, bytes.Equal(firstPrimary.Name.Buffer, secondPrimary.Name.Buffer), "Object Name doesn't match")
}

// TestCreate demonstrates that a non-storage parent can't create keys
// and that a storage parent can create keys
func TestCreate(t *testing.T) {
	tpm, err := simulator.OpenSimulator()
	if err != nil {
		log.Fatalf("can't open tpm simulator: %v", err)
	}
	defer tpm.Close()

	// not allowed to create keys
	signerPrimary, closer := setupCreatePrimary(t, tpm, tpm2.New2B(tpmutil.ECCSignerTemplate))
	defer closer()

	createCmd := tpm2.Create{
		ParentHandle: tpm2.NamedHandle{
			Name:   signerPrimary.Name,
			Handle: signerPrimary.ObjectHandle,
		},
		InPublic: tpm2.New2B(tpmutil.ECCSignerTemplate),
	}

	_, err = createCmd.Execute(tpm)
	assert.Error(t, err, "Non-Storage Parent should not be able to create keys")

	// allowed to create keys
	storageParentPrimary, storageParentCloser := setupCreatePrimary(t, tpm, tpmutil.ECCP256StorageParentTemplate)
	defer storageParentCloser()

	createCmd.ParentHandle = tpm2.NamedHandle{
		Name:   storageParentPrimary.Name,
		Handle: storageParentPrimary.ObjectHandle,
	}

	_, err = createCmd.Execute(tpm)
	assert.NoError(t, err, "Storage Parent should be able to create a key")
}
