package main

import (
	"crypto"
	"testing"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport"
	"github.com/stretchr/testify/require"
)

// setupCreatePrimary sets up the primary key creation and returns the response and a cleanup function.
func setupCreatePrimary(t *testing.T, tpm transport.TPM, public tpm2.TPM2BPublic) (*tpm2.CreatePrimaryResponse, func()) {
	createPrimaryRsp, err, closer := createPrimary(tpm, public)
	require.NoError(t, err, "CreatePrimary() failed")

	return createPrimaryRsp, closer
}

// mustGetEccPub is a helper function that extracts the ECC public key from the CreatePrimary response.
func mustGetEccPub(t *testing.T, primary *tpm2.CreatePrimaryResponse) crypto.PublicKey {
	pub, err := getEccPub(&primary.OutPublic)
	require.NoError(t, err, "getEccPub() failed")

	return pub
}
