package tpmutil

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport"
	"github.com/loicsikidi/go-tpm-kit/tpmcrypto"
	"github.com/loicsikidi/go-tpm-kit/tpmutil"
	"github.com/loicsikidi/tpm-pills/internal/keyutil"
	"github.com/loicsikidi/tpm-pills/internal/pemutil"
)

// CreatePrimary creates a simple primary key in the TPM and returns the response and a cleanup function.
func CreatePrimary(tpm transport.TPM, public tpm2.TPM2BPublic) (*tpm2.CreatePrimaryResponse, func(), error) {
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

// CreateOrdinaryKey creates an ordinary key in the TPM and saves the public and private keys to the specified directory.
// If the directory is empty, it uses the current working directory.
// Deprecated: Use [CreateKey] instead.
func CreateOrdinaryKey(tpm transport.TPM, outDir string, primaryTemplate, ordinaryTemplate tpm2.TPM2BPublic, createPublicKey bool) error {
	if outDir == "" {
		if dir, err := os.Getwd(); err == nil {
			outDir = dir
		}
	}
	primary, primaryCloser, err := CreatePrimary(tpm, primaryTemplate)
	if err != nil {
		return fmt.Errorf("failed to create primary key failed: %w", err)
	}
	defer primaryCloser()

	createRsp, err := tpm2.Create{
		ParentHandle: tpm2.NamedHandle{
			Name:   primary.Name,
			Handle: primary.ObjectHandle,
		},
		InPublic: ordinaryTemplate,
	}.Execute(tpm)
	if err != nil {
		return fmt.Errorf("failed to create ordinary key: %w", err)
	}

	// Save the TPM2B_PUBLIC
	if err := os.WriteFile(filepath.Join(outDir, "tpmkey.pub"), tpm2.Marshal(createRsp.OutPublic), 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}
	// Save the TPM2B_PRIVATE
	if err := os.WriteFile(filepath.Join(outDir, "tpmkey.priv"), tpm2.Marshal(createRsp.OutPrivate), 0644); err != nil {
		return fmt.Errorf("failed to write private blob: %w", err)
	}

	// Save the public key in PEM format
	if createPublicKey {
		pub, err := keyutil.PublicKey(&createRsp.OutPublic)
		if err != nil {
			return fmt.Errorf("failed to get public key: %w", err)
		}
		pem, err := pemutil.SerializePEMToBytes(pub)
		if err != nil {
			return fmt.Errorf("failed to serialize public key to PEM format: %w", err)
		}
		if err := os.WriteFile(filepath.Join(outDir, "public.pem"), pem, 0644); err != nil {
			return fmt.Errorf("failed to write public key: %w", err)
		}
	}
	return nil
}

// CreateKey creates an application key in the TPM using the provided configuration.
func CreateKey(tpm transport.TPM, cfg CreateKeyConfig) error {
	if err := cfg.CheckAndSetDefaults(); err != nil {
		return err
	}
	skrHandle, err := tpmutil.CreatePrimary(tpm, tpmutil.CreatePrimaryConfig{
		InPublic: cfg.ParentTemplate,
	})
	if err != nil {
		return fmt.Errorf("failed to create primary key failed: %w", err)
	}
	defer skrHandle.Close()

	createKeyResult, err := tpmutil.CreateWithResult(tpm, tpmutil.CreateConfig{
		ParentHandle: skrHandle,
		InPublic:     cfg.OrdinaryTemplate,
	})
	if err != nil {
		return fmt.Errorf("failed to create ordinary key: %w", err)
	}

	b, err := createKeyResult.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal create key result: %w", err)
	}

	if err := os.WriteFile(filepath.Join(cfg.OutDir, "key.tpm"), b, 0644); err != nil {
		return fmt.Errorf("failed to save tpm blob: %w", err)
	}

	if cfg.CreatePublicKey {
		if slices.Contains([]tpm2.TPMIAlgPublic{tpm2.TPMAlgECC, tpm2.TPMAlgRSA}, cfg.OrdinaryTemplate.Type) {
			pub, err := tpmcrypto.PublicKey(createKeyResult.PublicArea())
			if err != nil {
				return fmt.Errorf("failed to get public key: %w", err)
			}
			pem, err := pemutil.SerializePEMToBytes(pub)
			if err != nil {
				return fmt.Errorf("failed to serialize public key to PEM format: %w", err)
			}
			if err := os.WriteFile(filepath.Join(cfg.OutDir, "public.pem"), pem, 0644); err != nil {
				return fmt.Errorf("failed to write public key: %w", err)
			}
		}
	}
	return nil
}

// LoadOrdinaryKey loads an ordinary key into the TPM using the specified public and private key files.
// It creates a primary key using the provided template and loads the ordinary key into it.
// Deprecated: Use [LoadKey] instead.
func LoadOrdinaryKey(tpm transport.TPM, primaryTemplate tpm2.TPM2BPublic, pubPath, privPath string) (*tpm2.LoadResponse, func(), error) {
	pub, priv, err := loadTPMBlob(pubPath, privPath)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to get public and private keys: %w", err)
	}

	primary, primaryCloser, err := CreatePrimary(tpm, primaryTemplate)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to create primary key failed: %w", err)
	}
	defer primaryCloser()

	loadRsp, err := tpm2.Load{
		ParentHandle: tpm2.NamedHandle{
			Name:   primary.Name,
			Handle: primary.ObjectHandle,
		},
		InPublic:  *pub,
		InPrivate: *priv,
	}.Execute(tpm)
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to load ordinary key: %w", err)
	}
	flushContext := tpm2.FlushContext{FlushHandle: loadRsp.ObjectHandle}
	return loadRsp, func() {
		flushContext.Execute(tpm)
	}, nil
}

// loadTPMBlob loads the public and private keys from the specified files and unmarshals them into TPM2BPublic and TPM2BPrivate structures.
func loadTPMBlob(pubPath, privPath string) (*tpm2.TPM2BPublic, *tpm2.TPM2BPrivate, error) {
	b, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read public key: %w", err)
	}
	pub, err := tpm2.Unmarshal[tpm2.TPM2BPublic](b)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal TPM2BPublic: %w", err)
	}
	b, err = os.ReadFile(privPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read private key: %w", err)
	}
	priv, err := tpm2.Unmarshal[tpm2.TPM2BPrivate](b)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal TPM2BPrivate: %w", err)
	}
	return pub, priv, nil
}

func LoadKey(tpm transport.TPM, cfg LoadKeyConfig) (HandleCloser, error) {
	if err := cfg.CheckAndSetDefaults(); err != nil {
		return nil, err
	}
	skrHandle, err := tpmutil.CreatePrimary(tpm, tpmutil.CreatePrimaryConfig{
		InPublic: cfg.ParentTemplate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create primary key failed: %w", err)
	}
	defer skrHandle.Close()

	loadedBlob, err := tpmutil.LoadCreateResult(cfg.KeyBlobPath)
	if err != nil {
		return nil, err
	}
	return tpmutil.Load(tpm, tpmutil.LoadConfig{
		ParentHandle: skrHandle,
		InPublic:     loadedBlob.OutPublic,
		InPrivate:    loadedBlob.OutPrivate,
	})
}

func SymEncryptDecrypt(tpm transport.TPM, cfg SymEncryptDecryptConfig) ([]byte, error) {
	return tpmutil.SymEncryptDecrypt(tpm, tpmutil.SymEncryptDecryptConfig{
		KeyHandle: cfg.KeyHandle,
		Data:      cfg.Data,
		Mode:      cfg.Mode,
		IV:        cfg.IV,
		Decrypt:   cfg.Decrypt,
	})
}

func Seal(tpm transport.TPM, cfg SealConfig) error {
	if err := cfg.CheckAndSetDefaults(); err != nil {
		return err
	}
	skrHandle, err := tpmutil.CreatePrimary(tpm, tpmutil.CreatePrimaryConfig{
		InPublic: cfg.ParentTemplate,
	})
	if err != nil {
		return fmt.Errorf("failed to create primary key failed: %w", err)
	}
	defer skrHandle.Close()

	createKeyResult, err := tpmutil.CreateWithResult(tpm, tpmutil.CreateConfig{
		ParentHandle: skrHandle,
		InPublic:     SealTemplate,
		SealingData:  cfg.Message,
	})
	if err != nil {
		return fmt.Errorf("failed to seal data into TPM: %w", err)
	}

	b, err := createKeyResult.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal create key result: %w", err)
	}

	if err := os.WriteFile(cfg.OutputFilePath, b, 0644); err != nil {
		return fmt.Errorf("failed to save tpm blob: %w", err)
	}
	return nil
}

func Unseal(tpm transport.TPM, cfg UnsealConfig) ([]byte, error) {
	unsealRsp, err := tpm2.Unseal{
		ItemHandle: tpmutil.ToAuthHandle(cfg.KeyHandle),
	}.Execute(tpm)
	if err != nil {
		return nil, fmt.Errorf("failed to unseal data: %w", err)
	}
	return unsealRsp.OutData.Buffer, nil
}

func HMAC(tpm transport.TPM, cfg HMACConfig) ([]byte, error) {
	if err := cfg.CheckAndSetDefaults(); err != nil {
		return nil, err
	}
	hmacKeyHandle, err := tpmutil.CreatePrimary(tpm, tpmutil.CreatePrimaryConfig{
		InPublic: cfg.KeyTemplate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create primary key: %v", err)
	}
	defer hmacKeyHandle.Close()

	return tpmutil.Hmac(tpm, tpmutil.HmacConfig{
		KeyHandle: hmacKeyHandle,
		Data:      cfg.Data,
	})
}
