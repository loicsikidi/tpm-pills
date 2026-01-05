//go:build !windows

package main

import (
	"crypto/aes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport"
	"github.com/loicsikidi/tpm-pills/internal/options"
	"github.com/loicsikidi/tpm-pills/internal/tpmutil"
	"github.com/loicsikidi/tpm-pills/internal/utils"
)

var useTPM bool

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	createOpts := &options.CreateKeyOpts{
		KeyType: options.Decrypt.String(),
	}
	encryptOpts := &options.SymEncryptOpts{}
	decryptOpts := &options.DecryptOpts{}
	sealOpts := &options.SealOpts{}
	unsealOpts := &options.UnsealOpts{}
	hmacOpts := &options.HMACOpts{}

	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	encryptCmd := flag.NewFlagSet("encrypt", flag.ExitOnError)
	decryptCmd := flag.NewFlagSet("decrypt", flag.ExitOnError)

	// Define flags for the create subcommand
	createCmd.StringVar(&createOpts.OutputDir, "out", "", "Output directory for the created key")
	createCmd.BoolVar(&useTPM, "use-real-tpm", false, "Use real TPM instead of swtpm")

	// Define flags for the encrypt subcommand
	encryptCmd.StringVar(&encryptOpts.KeyBlobPath, "key", "", "Path to TPM key blob file")
	encryptCmd.StringVar(&encryptOpts.Message, "message", "", "Message to encrypt")
	encryptCmd.StringVar(&encryptOpts.OutputFilePath, "output", "", "Output file for the encrypted message")
	encryptCmd.BoolVar(&useTPM, "use-real-tpm", false, "Use real TPM instead of swtpm")

	// Define flags for the decrypt subcommand
	decryptCmd.StringVar(&decryptOpts.KeyBlobPath, "key", "", "Path to TPM key blob file")
	decryptCmd.StringVar(&decryptOpts.InputFilePath, "in", "", "Input file to decrypt")
	decryptCmd.BoolVar(&useTPM, "use-real-tpm", false, "Use real TPM instead of swtpm")

	// Define flags for the seal subcommand
	sealCmd := flag.NewFlagSet("seal", flag.ExitOnError)
	sealCmd.StringVar(&sealOpts.Message, "message", "", "Message to seal")
	sealCmd.StringVar(&sealOpts.OutputFilePath, "output", "", "Output file for the sealed message")
	sealCmd.BoolVar(&useTPM, "use-real-tpm", false, "Use real TPM instead of swtpm")

	// Define flags for the unseal subcommand
	unsealCmd := flag.NewFlagSet("unseal", flag.ExitOnError)
	unsealCmd.StringVar(&unsealOpts.InputFilePath, "in", "", "Input file to unseal")
	unsealCmd.BoolVar(&useTPM, "use-real-tpm", false, "Use real TPM instead of swtpm")

	hmacCmd := flag.NewFlagSet("hmac", flag.ExitOnError)
	hmacCmd.StringVar(&hmacOpts.Data, "data", "", "Data to compute HMAC for")
	hmacCmd.BoolVar(&useTPM, "use-real-tpm", false, "Use real TPM instead of swtpm")

	if len(os.Args) < 2 {
		flag.Usage()
		return fmt.Errorf("missing subcommand")
	}

	switch subcmd := os.Args[1]; subcmd {
	// commands involving a TPM
	case "create", "encrypt", "decrypt", "seal", "unseal", "hmac":
		switch subcmd {
		case "create":
			createCmd.Parse(os.Args[2:])
		case "decrypt":
			decryptCmd.Parse(os.Args[2:])
		case "encrypt":
			encryptCmd.Parse(os.Args[2:])
		case "seal":
			sealCmd.Parse(os.Args[2:])
		case "unseal":
			unsealCmd.Parse(os.Args[2:])
		case "hmac":
			hmacCmd.Parse(os.Args[2:])
		}

		var device tpmutil.Device
		if useTPM {
			device = tpmutil.LINUX
		} else {
			device = tpmutil.SWTPM
		}

		tpm, err := tpmutil.OpenTPM(device)
		if err != nil {
			return fmt.Errorf("can't open tpm: %w", err)
		}
		defer tpm.Close()

		if subcmd == "create" {
			if err := createCommand(tpm, createOpts); err != nil {
				return fmt.Errorf("error creating key: %w", err)
			}
			fmt.Println("Ordinary key created successfully 🚀")
		}
		if subcmd == "encrypt" {
			if err := encryptCommand(tpm, encryptOpts); err != nil {
				return fmt.Errorf("error encrypting message: %w", err)
			}
			fmt.Printf("Encrypted message saved to %s 🚀\n", encryptOpts.OutputFilePath)
		}
		if subcmd == "decrypt" {
			secret, err := decryptCommand(tpm, decryptOpts)
			if err != nil {
				return fmt.Errorf("error decrypting blob: %w", err)
			}
			fmt.Printf("Decrypted message: %q 🚀\n", secret)
		}
		if subcmd == "seal" {
			if err := sealCommand(tpm, sealOpts); err != nil {
				return fmt.Errorf("error sealing message: %w", err)
			}
			fmt.Printf("Sealed message saved to %s 🚀\n", sealOpts.OutputFilePath)
		}
		if subcmd == "unseal" {
			unsealedData, err := unsealCommand(tpm, unsealOpts)
			if err != nil {
				return fmt.Errorf("error unsealing message: %w", err)
			}
			fmt.Printf("Unsealed message: %q 🚀\n", string(unsealedData))
		}
		if subcmd == "hmac" {
			result, err := hmacCommand(tpm, hmacOpts)
			if err != nil {
				return fmt.Errorf("error computing HMAC: %w", err)
			}
			fmt.Printf("HMAC result: %q 🚀\n", hex.EncodeToString(result))

		}
	case "cleanup":
		if err := os.RemoveAll(tpmutil.SWTPM_ROOT_STATE); err != nil {
			return fmt.Errorf("error cleaning state: %w", err)
		}
		fmt.Println("State cleaned successfully 🚀")
	default:
		return fmt.Errorf("unknown subcommand %q. Expected 'create', 'encrypt', 'decrypt' or 'cleanup'", subcmd)
	}
	return nil
}

func createCommand(tpm transport.TPM, opts *options.CreateKeyOpts) error {
	if err := opts.CheckAndSetDefaults(); err != nil {
		return err
	}
	return tpmutil.CreateKey(tpm, tpmutil.CreateKeyConfig{
		OutDir:           opts.OutputDir,
		ParentTemplate:   tpmutil.ECCSRKTemplate,
		OrdinaryTemplate: tpmutil.SymTemplatesByKeyType[opts.GetKeyType()],
	})
}

type encryptedBlob struct {
	Ciphertext []byte `json:"ciphertext"`
	IV         []byte `json:"iv"`
}

func encryptCommand(tpm transport.TPM, opts *options.SymEncryptOpts) error {
	if err := opts.CheckAndSetDefaults(); err != nil {
		return err
	}

	keyHandle, err := tpmutil.LoadKey(tpm, tpmutil.LoadKeyConfig{
		ParentTemplate: tpmutil.ECCSRKTemplate,
		KeyBlobPath:    opts.KeyBlobPath,
	})
	if err != nil {
		return fmt.Errorf("error loading key blob: %v", err)
	}
	defer keyHandle.Close()

	iv := tpmutil.MustGenerateRnd(aes.BlockSize)
	ciphertext, err := tpmutil.SymEncryptDecrypt(tpm, tpmutil.SymEncryptDecryptConfig{
		KeyHandle: keyHandle,
		Data:      []byte(opts.Message),
		IV:        iv,
		Mode:      tpm2.TPMAlgCFB,
	})
	if err != nil {
		return fmt.Errorf("error encrypting message: %v", err)
	}

	blob, err := json.Marshal(encryptedBlob{
		Ciphertext: ciphertext,
		IV:         iv,
	})
	if err != nil {
		return fmt.Errorf("error marshaling encrypted blob: %v", err)
	}

	if err := os.WriteFile(opts.OutputFilePath, blob, 0644); err != nil {
		return fmt.Errorf("error writing encrypted blob to file: %v", err)
	}

	return nil
}

func decryptCommand(tpm transport.TPM, opts *options.DecryptOpts) ([]byte, error) {
	if err := opts.CheckAndSetDefaults(); err != nil {
		return nil, err
	}

	keyHandle, err := tpmutil.LoadKey(tpm, tpmutil.LoadKeyConfig{
		ParentTemplate: tpmutil.ECCSRKTemplate,
		KeyBlobPath:    opts.KeyBlobPath,
	})
	if err != nil {
		return nil, fmt.Errorf("error loading key blob: %v", err)
	}
	defer keyHandle.Close()

	var blob encryptedBlob
	// note: opts.CheckAndSetDefaults() ensures that InputFilePath exists
	data, _ := utils.ReadFile(opts.InputFilePath)
	if err := json.Unmarshal(data, &blob); err != nil {
		return nil, fmt.Errorf("error unmarshaling encrypted blob: %v", err)
	}

	return tpmutil.SymEncryptDecrypt(tpm, tpmutil.SymEncryptDecryptConfig{
		KeyHandle: keyHandle,
		Data:      blob.Ciphertext,
		IV:        blob.IV,
		Mode:      tpm2.TPMAlgCFB,
		Decrypt:   true,
	})
}

func sealCommand(tpm transport.TPM, opts *options.SealOpts) error {
	if err := opts.CheckAndSetDefaults(); err != nil {
		return err
	}

	return tpmutil.Seal(tpm, tpmutil.SealConfig{
		ParentTemplate: tpmutil.ECCSRKTemplate,
		Message:        []byte(opts.Message),
		OutputFilePath: opts.OutputFilePath,
	})
}

func unsealCommand(tpm transport.TPM, opts *options.UnsealOpts) ([]byte, error) {
	if err := opts.CheckAndSetDefaults(); err != nil {
		return nil, err
	}
	keyHandle, err := tpmutil.LoadKey(tpm, tpmutil.LoadKeyConfig{
		ParentTemplate: tpmutil.ECCSRKTemplate,
		KeyBlobPath:    opts.InputFilePath,
	})
	if err != nil {
		return nil, fmt.Errorf("error loading key blob: %v", err)
	}
	defer keyHandle.Close()

	unsealedData, err := tpmutil.Unseal(tpm, tpmutil.UnsealConfig{
		KeyHandle: keyHandle,
	})
	if err != nil {
		return nil, fmt.Errorf("error unsealing data: %v", err)
	}
	return unsealedData, nil
}

func hmacCommand(tpm transport.TPM, opts *options.HMACOpts) ([]byte, error) {
	if err := opts.CheckAndSetDefaults(); err != nil {
		return nil, err
	}

	hmacTemplate, err := tpmutil.NewHMACKeyTemplate(tpm2.TPMAlgSHA256)
	if err != nil {
		return nil, fmt.Errorf("error creating HMAC key template: %v", err)
	}

	return tpmutil.HMAC(tpm, tpmutil.HMACConfig{
		KeyTemplate: hmacTemplate,
		Data:        []byte(opts.Data),
	})
}
