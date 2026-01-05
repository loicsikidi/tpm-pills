package tpmutil

import (
	"fmt"

	"github.com/google/go-tpm/tpm2"
	"github.com/loicsikidi/tpm-pills/internal/utils"
)

type CreateKeyConfig struct {
	OutDir           string
	ParentTemplate   tpm2.TPMTPublic
	OrdinaryTemplate tpm2.TPMTPublic
	CreatePublicKey  bool
}

func (c *CreateKeyConfig) CheckAndSetDefaults() error {
	if c.OutDir == "" {
		dir, err := utils.FallbackDir()
		if err != nil {
			return err
		}
		c.OutDir = dir
	}
	return nil
}

type LoadKeyConfig struct {
	ParentTemplate tpm2.TPMTPublic
	KeyBlobPath    string
}

func (c *LoadKeyConfig) CheckAndSetDefaults() error {
	if c.KeyBlobPath == "" {
		return fmt.Errorf("invalid input: KeyBlobPath is required")
	}
	if !utils.FileExists(c.KeyBlobPath) {
		return fmt.Errorf("invalid input: KeyBlobPath does not exist")
	}
	return nil
}

type SymEncryptDecryptConfig struct {
	KeyHandle Handle
	Data      []byte
	Decrypt   bool
	IV        []byte
	Mode      tpm2.TPMAlgID
}

func (c *SymEncryptDecryptConfig) CheckAndSetDefaults() error {
	return nil
}

type SealConfig struct {
	ParentTemplate tpm2.TPMTPublic
	Message        []byte
	OutputFilePath string
}

func (c *SealConfig) CheckAndSetDefaults() error {
	if c.OutputFilePath == "" {
		return fmt.Errorf("invalid input: OutputFilePath is required")
	}
	if len(c.Message) == 0 {
		return fmt.Errorf("invalid input: Message is required")
	}
	return nil
}

type UnsealConfig struct {
	KeyHandle Handle
}

func (c *UnsealConfig) CheckAndSetDefaults() error {
	if c.KeyHandle == nil {
		return fmt.Errorf("invalid input: KeyHandle is required")
	}
	return nil
}

type HMACConfig struct {
	KeyTemplate tpm2.TPMTPublic
	Data        []byte
}

func (c *HMACConfig) CheckAndSetDefaults() error {
	if len(c.Data) == 0 {
		return fmt.Errorf("invalid input: Data is required")
	}
	return nil
}
