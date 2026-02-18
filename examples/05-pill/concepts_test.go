package main

import (
	"crypto/sha256"
	"testing"

	"github.com/google/go-tpm/tpm2"
	"github.com/loicsikidi/go-tpm-kit/tpmtest"
	"github.com/loicsikidi/tpm-pills/internal/tpmutil"
)

func TestSignExternalDataWithRestrictedKey(t *testing.T) {
	type digestAndTicket struct {
		Digest tpm2.TPM2BDigest
		Ticket tpm2.TPMTTKHashCheck
	}

	message := []byte("test message")

	tpm := tpmtest.OpenSimulator(t)

	tests := []struct {
		name            string
		digestAndTicket digestAndTicket
		wantErr         bool
	}{
		{
			name: "ok: hash produced by TPM",
			digestAndTicket: func() digestAndTicket {
				rspHash, err := tpm2.Hash{
					Data:      tpm2.TPM2BMaxBuffer{Buffer: message},
					HashAlg:   tpm2.TPMAlgSHA256,
					Hierarchy: tpm2.TPMRHOwner,
				}.Execute(tpm)
				if err != nil {
					t.Fatalf("failed to execute hash command: %v", err)
				}
				return digestAndTicket{
					Digest: rspHash.OutHash,
					Ticket: rspHash.Validation,
				}
			}(),
			wantErr: false,
		},
		{
			name: "failure: hash not produced by TPM",
			digestAndTicket: func() digestAndTicket {
				msgDigest := sha256.Sum256([]byte(message))
				return digestAndTicket{
					Digest: tpm2.TPM2BDigest{
						Buffer: msgDigest[:],
					},
					Ticket: tpm2.TPMTTKHashCheck{
						Tag: tpm2.TPMSTHashCheck,
					},
				}
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			srkCreate, err := tpm2.CreatePrimary{
				PrimaryHandle: tpm2.TPMRHOwner,
				InPublic:      tpm2.New2B(tpm2.ECCSRKTemplate),
			}.Execute(tpm)
			if err != nil {
				t.Fatalf("could not create SRK: %v", err)
			}

			defer tpm2.FlushContext{
				FlushHandle: srkCreate.ObjectHandle,
			}.Execute(tpm)

			create, err := tpm2.Create{
				ParentHandle: tpm2.NamedHandle{
					Handle: srkCreate.ObjectHandle,
					Name:   srkCreate.Name,
				},
				InPublic: tpm2.New2B(tpmutil.ECCRestrictedSignerTemplate),
			}.Execute(tpm)
			if err != nil {
				t.Fatalf("could not create key: %v", err)
			}

			load, err := tpm2.Load{
				ParentHandle: tpm2.NamedHandle{
					Handle: srkCreate.ObjectHandle,
					Name:   srkCreate.Name,
				},
				InPrivate: create.OutPrivate,
				InPublic:  create.OutPublic,
			}.Execute(tpm)
			if err != nil {
				t.Fatalf("could not load key: %v", err)
			}
			defer tpm2.FlushContext{
				FlushHandle: load.ObjectHandle,
			}.Execute(tpm)

			_, err = tpm2.Sign{
				KeyHandle: tpm2.NamedHandle{
					Handle: load.ObjectHandle,
					Name:   load.Name,
				},
				Digest:     tt.digestAndTicket.Digest,
				Validation: tt.digestAndTicket.Ticket,
			}.Execute(tpm)

			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
