package main

import (
	"crypto"
	"testing"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport"
	"github.com/loicsikidi/tpm-pills/internal/keyutil"
	"github.com/loicsikidi/tpm-pills/internal/tpmutil"
	"github.com/stretchr/testify/require"
)

// setupCreatePrimary sets up the primary key creation and returns the response and a cleanup function.
func setupCreatePrimary(t *testing.T, tpm transport.TPM, public tpm2.TPM2BPublic) (*tpm2.CreatePrimaryResponse, func()) {
	createPrimaryRsp, closer, err := tpmutil.CreatePrimary(tpm, public)
	require.NoError(t, err, "CreatePrimary() failed")

	return createPrimaryRsp, closer
}

// mustPublicKey is a helper function that extracts the public key from the CreatePrimary response.
func mustPublicKey(t *testing.T, primary *tpm2.CreatePrimaryResponse) crypto.PublicKey {
	pub, err := keyutil.PublicKey(&primary.OutPublic)
	require.NoError(t, err, "keyutil.PublicKey() failed")

	return pub
}
