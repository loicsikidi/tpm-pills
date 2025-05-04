package main

import (
	"crypto"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport"
)

var (
	ECCSignerTemplate = tpm2.New2B(tpm2.TPMTPublic{
		Type:    tpm2.TPMAlgECC,
		NameAlg: tpm2.TPMAlgSHA256,
		ObjectAttributes: tpm2.TPMAObject{
			FixedTPM:            true,
			FixedParent:         true,
			SensitiveDataOrigin: true,
			UserWithAuth:        true,
			SignEncrypt:         true,
		},
		Parameters: tpm2.NewTPMUPublicParms(
			tpm2.TPMAlgECC,
			&tpm2.TPMSECCParms{
				Scheme: tpm2.TPMTECCScheme{
					Scheme: tpm2.TPMAlgECDSA,
					Details: tpm2.NewTPMUAsymScheme(
						tpm2.TPMAlgECDSA,
						&tpm2.TPMSSigSchemeECDSA{
							HashAlg: tpm2.TPMAlgSHA256,
						},
					),
				},
				CurveID: tpm2.TPMECCNistP256,
			},
		),
	})
	ECCStorageParentTemplate = tpm2.New2B(tpm2.TPMTPublic{
		Type:    tpm2.TPMAlgECC,
		NameAlg: tpm2.TPMAlgSHA256,
		ObjectAttributes: tpm2.TPMAObject{
			FixedTPM:            true,
			FixedParent:         true,
			SensitiveDataOrigin: true,
			UserWithAuth:        true,
			Restricted:          true,
			Decrypt:             true,
			SignEncrypt:         false,
		},
		Parameters: tpm2.NewTPMUPublicParms(
			tpm2.TPMAlgECC,
			&tpm2.TPMSECCParms{
				Symmetric: tpm2.TPMTSymDefObject{
					Algorithm: tpm2.TPMAlgAES,
					KeyBits: tpm2.NewTPMUSymKeyBits(
						tpm2.TPMAlgAES,
						tpm2.TPMKeyBits(128),
					),
					Mode: tpm2.NewTPMUSymMode(
						tpm2.TPMAlgAES,
						tpm2.TPMAlgCFB,
					),
				},
				CurveID: tpm2.TPMECCNistP256,
			},
		),
	})
)

// createPrimary creates a primary key in the TPM and returns the response and a cleanup function.
func createPrimary(tpm transport.TPM, public tpm2.TPM2BPublic) (*tpm2.CreatePrimaryResponse, func(), error) {
	createPrimaryCmd := tpm2.CreatePrimary{
		PrimaryHandle: tpm2.TPMRHOwner,
		InPublic:      public,
	}
	createPrimaryRsp, err := createPrimaryCmd.Execute(tpm)
	if err == nil {
		flushContext := tpm2.FlushContext{FlushHandle: createPrimaryRsp.ObjectHandle}
		return createPrimaryRsp, func() {
			flushContext.Execute(tpm)
		}, nil
	}
	return nil, func() {}, err
}

// getECCPub extracts the ECC public key from TPM2B_PUBLIC structure.
func getEccPub(tpm2public *tpm2.TPM2BPublic) (crypto.PublicKey, error) {
	pub, err := tpm2public.Contents()
	if err != nil {
		return nil, err
	}

	eccDetail, err := pub.Parameters.ECCDetail()
	if err != nil {
		return nil, err
	}

	eccUnique, err := pub.Unique.ECC()
	if err != nil {
		return nil, err
	}
	eccPub, err := tpm2.ECDSAPub(eccDetail, eccUnique)
	if err != nil {
		return nil, err
	}
	return eccPub, nil
}
