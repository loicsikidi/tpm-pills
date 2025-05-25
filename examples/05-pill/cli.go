//go:build !windows

package main

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/asn1"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport"
	"github.com/loicsikidi/tpm-pills/internal/pemutil"
	"github.com/loicsikidi/tpm-pills/internal/tpmutil"
)

var useTPM = flag.Bool("use-real-tpm", false, "Use real TPM instead of swtpm")

func main() {
	flag.Parse()

	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	encryptCmd := flag.NewFlagSet("encrypt", flag.ExitOnError)
	decryptCmd := flag.NewFlagSet("decrypt", flag.ExitOnError)
	signCmd := flag.NewFlagSet("sign", flag.ExitOnError)
	verifyCmd := flag.NewFlagSet("verify", flag.ExitOnError)

	// Define flags for the create subcommand
	createOutDir := createCmd.String("out", "", "Output directory for the created key")
	createKeyType := createCmd.String("type", "decrypt", "Key type to create (decrypt, signer or restrictedSigner)")

	// Define flags for the encrypt subcommand
	encryptKey := encryptCmd.String("key", "", "Path to the public key file")
	encryptMsg := encryptCmd.String("message", "", "Message to encrypt")
	encryptOut := encryptCmd.String("output", "", "Output file for the encrypted message")

	// Define flags for the decrypt subcommand
	public := decryptCmd.String("public", "", "Path to TPM public file")
	private := decryptCmd.String("private", "", "Path to TPM private file")
	in := decryptCmd.String("in", "", "Input file to decrypt")

	// Define flags for the sign subcommand
	signMsg := signCmd.String("message", "", "Message to sign")
	signOut := signCmd.String("output", "", "Output file for the signed message")
	signPub := signCmd.String("public", "", "Path to TPM public file")
	signPriv := signCmd.String("private", "", "Path to TPM private file")

	// Define flags for the verify subcommand
	verifyMsg := verifyCmd.String("message", "", "Message to verify")
	verifySig := verifyCmd.String("signature", "", "Path to the signature file")
	verifyKey := verifyCmd.String("key", "", "Path to the public key file")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'create' or 'decrypt' subcommands")
		os.Exit(1)
	}

	var device tpmutil.Device
	if *useTPM {
		device = tpmutil.LINUX
	} else {
		device = tpmutil.SWTPM
	}

	switch subcmd := os.Args[1]; subcmd {
	// commands involving a TPM
	case "create", "decrypt", "sign":
		tpm, err := tpmutil.OpenTPM(device)
		if err != nil {
			log.Fatalf("can't open tpm: %v", err)
		}
		defer tpm.Close()

		if subcmd == "create" {
			createCmd.Parse(os.Args[2:])
			if err := createCommand(tpm, *createKeyType, *createOutDir); err != nil {
				fmt.Printf("Error creating key: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Ordinary key created successfully 🚀")
		}
		if subcmd == "decrypt" {
			decryptCmd.Parse(os.Args[2:])
			secret, err := decryptCommand(tpm, *in, *public, *private)
			if err != nil {
				fmt.Printf("Error decrypting blob: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Decrypted %q successfully 🚀\n", secret)
		}
		if subcmd == "sign" {
			signCmd.Parse(os.Args[2:])
			if err := signCommand(tpm, *signMsg, *signOut, *signPub, *signPriv); err != nil {
				fmt.Printf("Error signing message: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Signature saved to %s 🚀\n", *signOut)
		}
	case "encrypt":
		encryptCmd.Parse(os.Args[2:])
		if err := encryptCommand(*encryptKey, *encryptMsg, *encryptOut); err != nil {
			fmt.Printf("Error encrypting message: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Encrypted message saved to %s 🚀\n", *encryptOut)
	case "verify":
		verifyCmd.Parse(os.Args[2:])
		if err := verifyCommand(*verifyMsg, *verifySig, *verifyKey); err != nil {
			fmt.Printf("Error verifying signature: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Signature verified successfully 🚀")
	case "cleanup":
		if err := os.RemoveAll(tpmutil.SWTPM_ROOT_STATE); err != nil {
			fmt.Printf("Error cleaning state: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("State cleaned successfully 🚀")
	default:
		fmt.Println("Unknown subcommand. Expected 'create', 'encrypt', 'decrypt', 'sign' or 'cleanup'")
		os.Exit(1)
	}
}

func createCommand(tpm transport.TPM, keyType, outDir string) error {
	var template tpm2.TPM2BPublic
	switch keyType {
	case "decrypt":
		template = tpmutil.RSA2048EncryptTemplate
	case "signer":
		template = tpmutil.ECCP256SignerTemplate
	case "restrictedSigner":
		template = tpmutil.ECCP256RestrictedSignerTemplate
	default:
		return fmt.Errorf("unknown key type %q. Expected 'decrypt', 'signer' or 'restrictedSigner'", keyType)
	}
	if err := tpmutil.CreateOrdinaryKey(tpm,
		outDir,
		tpmutil.ECCP256StorageParentTemplate,
		template,
		/* createPublicKey */ true); err != nil {
		return err
	}
	return nil
}

func encryptCommand(keyPath, message, outPath string) error {
	if keyPath == "" || message == "" || outPath == "" {
		return fmt.Errorf("--key, --message and --output flags are required for 'encrypt'")
	}
	pubKey, err := pemutil.Read(keyPath)
	if err != nil {
		return fmt.Errorf("error reading public key: %v", err)
	}
	rsaKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("error converting public key to RSA public key: %v", err)
	}
	ciphertext, err := encryptBlob(rsaKey, []byte(message))
	if err != nil {
		return err
	}
	if err := writeFile(ciphertext, outPath); err != nil {
		return err
	}
	return nil
}

func decryptCommand(tpm transport.TPM, inPath, pubPath, privPath string) ([]byte, error) {
	if pubPath == "" || privPath == "" || inPath == "" {
		return nil, fmt.Errorf("--public, --private and --in flags are required for 'decrypt'")
	}
	secret, err := decryptBlob(tpm,
		tpmutil.ECCP256StorageParentTemplate,
		inPath,
		pubPath,
		privPath,
	)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func signCommand(tpm transport.TPM, message, outPath, pubPath, privPath string) error {
	if pubPath == "" || privPath == "" || message == "" || outPath == "" {
		return fmt.Errorf("--public, --private, --message and --output flags are required for 'sign'")
	}
	signature, err := signBlob(tpm,
		tpmutil.ECCP256StorageParentTemplate,
		message,
		pubPath,
		privPath,
	)
	if err != nil {
		return err
	}
	if err := writeFile(signature, outPath); err != nil {
		return err
	}
	return nil
}

func verifyCommand(message, signaturePath, keyPath string) error {
	if keyPath == "" || message == "" || signaturePath == "" {
		return fmt.Errorf("--key, --message and --signature flags are required for 'verify'")
	}
	msgDigest := sha256.Sum256([]byte(message))

	pubKey, err := pemutil.Read(keyPath)
	if err != nil {
		return fmt.Errorf("error reading public key: %v", err)
	}
	ecdsaKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error converting public key to ECDSA public key: %v", err)
	}

	sig, err := os.ReadFile(signaturePath)
	if err != nil {
		return fmt.Errorf("error reading signature file: %v", err)
	}
	ecdsaSig := new(ecdsaSignature)
	_, err = asn1.Unmarshal(sig, ecdsaSig)
	if err != nil {
		return fmt.Errorf("error unmarshalling signature: %v", err)
	}
	valid := ecdsa.Verify(ecdsaKey, msgDigest[:], ecdsaSig.R, ecdsaSig.S)
	if !valid {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}
