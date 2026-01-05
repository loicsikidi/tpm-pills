//go:build !windows

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"math/big"
	"os"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport"
	"github.com/loicsikidi/tpm-pills/internal/tpmutil"
	"github.com/loicsikidi/tpm-pills/internal/utils"
)

type ecdsaSignature struct {
	R, S *big.Int
}

func decryptBlob(tpm transport.TPM, primaryTemplate tpm2.TPMTPublic, inPath, keyBlobPath string) ([]byte, error) {
	keyHandle, err := tpmutil.LoadKey(tpm, tpmutil.LoadKeyConfig{
		ParentTemplate: primaryTemplate,
		KeyBlobPath:    keyBlobPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load key: %w", err)
	}
	defer keyHandle.Close()

	ciphertext, _ := utils.ReadFile(inPath)

	decryptCmd := tpm2.RSADecrypt{
		KeyHandle:  keyHandle,
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
		return nil, fmt.Errorf("failed to execute decrypt command: %w", err)
	}
	return decryptRsp.Message.Buffer, nil
}

func writeFile(content []byte, outPath string) error {
	if err := os.WriteFile(outPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", outPath, err)
	}
	return nil
}

func encryptBlob(pub *rsa.PublicKey, message []byte) ([]byte, error) {
	ciphertext, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		pub,
		message,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt message: %w", err)
	}
	return ciphertext, nil
}

func signBlob(tpm transport.TPM, primaryTemplate tpm2.TPMTPublic, message, keyBlobPath string) ([]byte, error) {
	keyHandle, err := tpmutil.LoadKey(tpm, tpmutil.LoadKeyConfig{
		ParentTemplate: primaryTemplate,
		KeyBlobPath:    keyBlobPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load key: %w", err)
	}
	defer keyHandle.Close()

	if !keyHandle.HasPublic() {
		return nil, fmt.Errorf("key handle does not have a public key")
	}

	var (
		digest     tpm2.TPM2BDigest
		validation tpm2.TPMTTKHashCheck
	)
	if keyHandle.Public().ObjectAttributes.Restricted {
		rspHash, err := tpm2.Hash{
			Data:      tpm2.TPM2BMaxBuffer{Buffer: []byte(message)},
			HashAlg:   tpm2.TPMAlgSHA256,
			Hierarchy: tpm2.TPMRHOwner,
		}.Execute(tpm)
		if err != nil {
			return nil, fmt.Errorf("failed to execute hash command: %w", err)
		}
		digest = rspHash.OutHash
		validation = rspHash.Validation
	} else {
		msgDigest := sha256.Sum256([]byte(message))
		digest = tpm2.TPM2BDigest{
			Buffer: msgDigest[:],
		}
		// NULL ticket
		validation = tpm2.TPMTTKHashCheck{
			Tag:       tpm2.TPMSTHashCheck,
			Hierarchy: tpm2.TPMRHNull,
		}
	}

	signRsp, err := tpm2.Sign{
		KeyHandle:  keyHandle,
		Digest:     digest,
		Validation: validation,
	}.Execute(tpm)

	if err != nil {
		return nil, fmt.Errorf("failed to execute sign command: %w", err)
	}

	eccSig, err := signRsp.Signature.Signature.ECDSA()
	if err != nil {
		return nil, fmt.Errorf("failed to get ECDSA signature: %w", err)
	}

	r := new(big.Int).SetBytes(eccSig.SignatureR.Buffer)
	s := new(big.Int).SetBytes(eccSig.SignatureS.Buffer)

	sig := ecdsaSignature{R: r, S: s}
	der, err := asn1.Marshal(sig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ECDSA signature: %w", err)
	}
	return der, nil
}
