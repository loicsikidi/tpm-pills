package pemutil

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func SerializePEM(in any) (*pem.Block, error) {
	var p *pem.Block
	switch k := in.(type) {
	case *rsa.PublicKey, *ecdsa.PublicKey:
		b, err := x509.MarshalPKIXPublicKey(k)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal public key: %w", err)
		}
		p = &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: b,
		}
	default:
		return nil, fmt.Errorf("cannot serialize type '%T', value '%v'", k, k)
	}
	return p, nil
}

func SerializePEMToBytes(in any) ([]byte, error) {
	p, err := SerializePEM(in)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(p), nil
}

// Parse returns the key or certificate PEM-encoded in the given bytes.
func Parse(b []byte) (any, error) {
	block, rest := pem.Decode(b)
	switch {
	case block == nil:
		return nil, fmt.Errorf("error decoding: not a valid PEM encoded block")
	case len(bytes.TrimSpace(rest)) > 0:
		return nil, fmt.Errorf("error decoding: contains more than one PEM encoded block")
	}

	switch block.Type {
	case "PUBLIC KEY":
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing public key: %w", err)
		}
		return pub, nil
	default:
		return nil, fmt.Errorf("error decoding: contains an unexpected header %q", block.Type)
	}
}

func Read(filename string) (any, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return Parse(b)
}
