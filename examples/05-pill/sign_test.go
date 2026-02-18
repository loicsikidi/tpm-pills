package main

import (
	"testing"

	"github.com/google/go-tpm/tpm2"
	"github.com/loicsikidi/go-tpm-kit/tpmtest"
	"github.com/loicsikidi/tpm-pills/internal/tpmutil"
)

// The creation of the ticket may be suppressed by using TPM_RH_NULL
// as the hierarchy parameter in TPM2_Hash() or TPM2_SequenceComplete()
func TestHashHierarchyWithRestrictedKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		hierarchy     tpm2.TPMHandle
		hashHierarchy tpm2.TPMHandle
		wantErr       bool
	}{
		{
			name:          "ok [key: owner | hash: owner]",
			hierarchy:     tpm2.TPMRHOwner,
			hashHierarchy: tpm2.TPMRHOwner,
			wantErr:       false,
		},
		{
			name:          "ok [key: owner | hash: endorsement]",
			hierarchy:     tpm2.TPMRHOwner,
			hashHierarchy: tpm2.TPMRHEndorsement,
			wantErr:       false,
		},
		{
			name:          "ok [key: owner | hash: platform]",
			hierarchy:     tpm2.TPMRHOwner,
			hashHierarchy: tpm2.TPMRHPlatform,
			wantErr:       false,
		},
		{
			name:          "ok [key: endorsement | hash: endorsement]",
			hierarchy:     tpm2.TPMRHEndorsement,
			hashHierarchy: tpm2.TPMRHEndorsement,
			wantErr:       false,
		},
		{
			name:          "ko [key: owner | hash: null]",
			hierarchy:     tpm2.TPMRHOwner,
			hashHierarchy: tpm2.TPMRHNull,
			wantErr:       true,
		},
		{
			name:          "ko [key: null | hash: null]",
			hierarchy:     tpm2.TPMRHNull,
			hashHierarchy: tpm2.TPMRHNull,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpm := tpmtest.OpenSimulator(t)

			srkCreate, err := tpm2.CreatePrimary{
				PrimaryHandle: tt.hierarchy,
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

			rspHash, err := tpm2.Hash{
				Data:      tpm2.TPM2BMaxBuffer{Buffer: []byte("test message")},
				HashAlg:   tpm2.TPMAlgSHA256,
				Hierarchy: tt.hashHierarchy,
			}.Execute(tpm)
			if err != nil {
				t.Fatalf("failed to execute hash command: %v", err)
			}

			_, err = tpm2.Sign{
				KeyHandle: tpm2.NamedHandle{
					Handle: load.ObjectHandle,
					Name:   load.Name,
				},
				Digest:     rspHash.OutHash,
				Validation: rspHash.Validation,
			}.Execute(tpm)

			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
