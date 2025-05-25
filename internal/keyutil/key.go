package keyutil

import (
	"crypto"
	"fmt"

	"github.com/google/go-tpm/tpm2"
)

func PublicKey(tpm2public *tpm2.TPM2BPublic) (crypto.PublicKey, error) {
	pub, err := tpm2public.Contents()
	if err != nil {
		return nil, err
	}

	switch pub.Type {
	case tpm2.TPMAlgRSA:
		rsaDetail, err := pub.Parameters.RSADetail()
		if err != nil {
			return nil, err
		}
		rsaUnique, err := pub.Unique.RSA()
		if err != nil {
			return nil, err
		}
		rsaPub, err := tpm2.RSAPub(rsaDetail, rsaUnique)
		if err != nil {
			return nil, err
		}
		return rsaPub, nil
	case tpm2.TPMAlgECC:
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
	default:
		return nil, fmt.Errorf("unrecognized key type: %T", pub.Type)
	}
}
