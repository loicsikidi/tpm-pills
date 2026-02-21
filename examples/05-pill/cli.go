//go:build !windows

package main

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/asn1"
	"flag"
	"fmt"
	"os"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport"
	"github.com/loicsikidi/tpm-pills/internal/options"
	"github.com/loicsikidi/tpm-pills/internal/pemutil"
	"github.com/loicsikidi/tpm-pills/internal/tpmutil"
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
	encryptOpts := &options.EncryptOpts{}
	decryptOpts := &options.AsymDecryptOpts{}
	signOpts := &options.SignOpts{}
	verifyOpts := &options.VerifyOpts{}

	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	encryptCmd := flag.NewFlagSet("encrypt", flag.ExitOnError)
	decryptCmd := flag.NewFlagSet("decrypt", flag.ExitOnError)
	signCmd := flag.NewFlagSet("sign", flag.ExitOnError)
	verifyCmd := flag.NewFlagSet("verify", flag.ExitOnError)

	// Define flags for the create subcommand
	createCmd.StringVar(&createOpts.OutputDir, "out", "", "Output directory for the created key")
	createCmd.StringVar(&createOpts.KeyType, "type", "decrypt", "Key type to create (decrypt, signer or restrictedSigner)")
	createCmd.BoolVar(&useTPM, "use-real-tpm", false, "Use real TPM instead of swtpm")

	// Define flags for the encrypt subcommand
	encryptCmd.StringVar(&encryptOpts.PublicKeyPath, "pubkey", "", "Path to the public key file")
	encryptCmd.StringVar(&encryptOpts.Message, "message", "", "Message to encrypt")
	encryptCmd.StringVar(&encryptOpts.OutputFilePath, "output", "", "Output file for the encrypted message")
	encryptCmd.BoolVar(&useTPM, "use-real-tpm", false, "Use real TPM instead of swtpm")

	// Define flags for the decrypt subcommand
	decryptCmd.StringVar(&decryptOpts.KeyBlobPath, "key", "", "Path to TPM key blob file")
	decryptCmd.StringVar(&decryptOpts.InputFilePath, "in", "", "Input file to decrypt")
	decryptCmd.BoolVar(&useTPM, "use-real-tpm", false, "Use real TPM instead of swtpm")

	// Define flags for the sign subcommand
	signCmd.StringVar(&signOpts.KeyBlobPath, "key", "", "Path to TPM key blob file")
	signCmd.StringVar(&signOpts.Message, "message", "", "Message to sign")
	signCmd.StringVar(&signOpts.OutputFilePath, "output", "", "Output file for the signed message")
	signCmd.BoolVar(&useTPM, "use-real-tpm", false, "Use real TPM instead of swtpm")

	// Define flags for the verify subcommand
	verifyCmd.StringVar(&verifyOpts.PublicKeyPath, "pubkey", "", "Path to the public key file")
	verifyCmd.StringVar(&verifyOpts.Message, "message", "", "Message to verify")
	verifyCmd.StringVar(&verifyOpts.SignaturePath, "signature", "", "Path to the signature file")
	verifyCmd.BoolVar(&useTPM, "use-real-tpm", false, "Use real TPM instead of swtpm")

	if len(os.Args) < 2 {
		flag.Usage()
		return fmt.Errorf("missing subcommand")
	}

	switch subcmd := os.Args[1]; subcmd {
	// commands involving a TPM
	case "create", "decrypt", "sign":
		switch subcmd {
		case "create":
			createCmd.Parse(os.Args[2:])
		case "decrypt":
			decryptCmd.Parse(os.Args[2:])
		case "sign":
			signCmd.Parse(os.Args[2:])
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
		if subcmd == "decrypt" {
			secret, err := decryptCommand(tpm, decryptOpts)
			if err != nil {
				return fmt.Errorf("error decrypting blob: %w", err)
			}
			fmt.Printf("Decrypted %q successfully 🚀\n", secret)
		}
		if subcmd == "sign" {
			if err := signCommand(tpm, signOpts); err != nil {
				return fmt.Errorf("error signing message: %w", err)
			}
			fmt.Printf("Signature saved to %s 🚀\n", signOpts.OutputFilePath)
		}
	case "encrypt":
		encryptCmd.Parse(os.Args[2:])
		if err := encryptCommand(encryptOpts); err != nil {
			return fmt.Errorf("error encrypting message: %w", err)
		}
		fmt.Printf("Encrypted message saved to %s 🚀\n", encryptOpts.OutputFilePath)
	case "verify":
		verifyCmd.Parse(os.Args[2:])
		if err := verifyCommand(verifyOpts); err != nil {
			return fmt.Errorf("error verifying signature: %w", err)
		}
		fmt.Println("Signature verified successfully 🚀")
	case "cleanup":
		if err := os.RemoveAll(tpmutil.SWTPM_ROOT_STATE); err != nil {
			return fmt.Errorf("error cleaning state: %w", err)
		}
		fmt.Println("State cleaned successfully 🚀")
	default:
		return fmt.Errorf("unknown subcommand %q. Expected 'create', 'encrypt', 'decrypt', 'sign', 'verify' or 'cleanup'", subcmd)
	}
	return nil
}

func createCommand(tpm transport.TPM, opts *options.CreateKeyOpts) error {
	if err := opts.CheckAndSetDefaults(); err != nil {
		return err
	}

	var template tpm2.TPMTPublic
	switch opts.GetKeyType() {
	case options.Decrypt:
		template = tpmutil.RSAEncryptTemplate
	case options.Signer:
		template = tpmutil.ECCSignerTemplate
	case options.RestrictedSigner:
		template = tpmutil.ECCRestrictedSignerTemplate
	default:
		return fmt.Errorf("unknown key type %q. Expected 'decrypt', 'signer' or 'restricted-signer'", opts.GetKeyType())
	}

	return tpmutil.CreateKey(tpm, tpmutil.CreateKeyConfig{
		OutDir:           opts.OutputDir,
		ParentTemplate:   tpmutil.ECCSRKTemplate,
		OrdinaryTemplate: template,
		CreatePublicKey:  true,
	})
}

func encryptCommand(opts *options.EncryptOpts) error {
	if err := opts.CheckAndSetDefaults(); err != nil {
		return err
	}

	pubKey, err := pemutil.Read(opts.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("error reading public key: %w", err)
	}
	rsaKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("error converting public key to RSA public key")
	}
	ciphertext, err := encryptBlob(rsaKey, []byte(opts.Message))
	if err != nil {
		return err
	}
	if err := writeFile(ciphertext, opts.OutputFilePath); err != nil {
		return err
	}
	return nil
}

func decryptCommand(tpm transport.TPM, opts *options.AsymDecryptOpts) ([]byte, error) {
	if err := opts.CheckAndSetDefaults(); err != nil {
		return nil, err
	}

	return decryptBlob(tpm,
		tpmutil.ECCSRKTemplate,
		opts.InputFilePath,
		opts.KeyBlobPath,
	)
}

func signCommand(tpm transport.TPM, opts *options.SignOpts) error {
	if err := opts.CheckAndSetDefaults(); err != nil {
		return err
	}

	signature, err := signBlob(tpm,
		tpmutil.ECCSRKTemplate,
		opts.Message,
		opts.KeyBlobPath,
	)
	if err != nil {
		return err
	}
	if err := writeFile(signature, opts.OutputFilePath); err != nil {
		return err
	}
	return nil
}

func verifyCommand(opts *options.VerifyOpts) error {
	if err := opts.CheckAndSetDefaults(); err != nil {
		return err
	}

	msgDigest := sha256.Sum256([]byte(opts.Message))

	pubKey, err := pemutil.Read(opts.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("error reading public key: %w", err)
	}
	ecdsaKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error converting public key to ECDSA public key")
	}

	sig, err := os.ReadFile(opts.SignaturePath)
	if err != nil {
		return fmt.Errorf("error reading signature file: %w", err)
	}
	ecdsaSig := new(ecdsaSignature)
	_, err = asn1.Unmarshal(sig, ecdsaSig)
	if err != nil {
		return fmt.Errorf("error unmarshalling signature: %w", err)
	}
	valid := ecdsa.Verify(ecdsaKey, msgDigest[:], ecdsaSig.R, ecdsaSig.S)
	if !valid {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}
