package main

import (
	"bytes"
	"log"
	"testing"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport/simulator"
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

	firstPrimary, closer := setupCreatePrimary(t, tpm, ECCSignerTemplate)
	defer closer()

	secondPrimary, secondCloser := setupCreatePrimary(t, tpm, ECCSignerTemplate)
	defer secondCloser()

	firstEccPub := mustGetEccPub(t, firstPrimary)
	secondEccPub := mustGetEccPub(t, secondPrimary)

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

	signerPrimary, closer := setupCreatePrimary(t, tpm, ECCSignerTemplate)
	defer closer()

	createCmd := tpm2.Create{
		ParentHandle: tpm2.NamedHandle{
			Name:   signerPrimary.Name,
			Handle: signerPrimary.ObjectHandle,
		},
		InPublic: ECCSignerTemplate,
	}

	_, err = createCmd.Execute(tpm)
	assert.Error(t, err, "Non-Storage Parent should not be able to create keys")

	storageParentPrimary, storageParentCloser := setupCreatePrimary(t, tpm, ECCStorageParentTemplate)
	defer storageParentCloser()

	createCmd.ParentHandle = tpm2.NamedHandle{
		Name:   storageParentPrimary.Name,
		Handle: storageParentPrimary.ObjectHandle,
	}

	_, err = createCmd.Execute(tpm)
	assert.NoError(t, err, "Storage Parent should be able to create a key")
}
