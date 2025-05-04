//go:build !windows

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport"
	"github.com/loicsikidi/tpm-pills/internal/tpmutil"
)

var useTPM = flag.Bool("use-real-tpm", false, "Use real TPM instead of swtpm")

func main() {
	flag.Parse()

	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	loadCmd := flag.NewFlagSet("load", flag.ExitOnError)

	// Define flags for the create subcommand
	createOutDir := createCmd.String("out", "", "Output directory for the created key")

	// Define flags for the load subcommand
	loadPub := loadCmd.String("public", "", "Path to the public key file")
	loadPrivate := loadCmd.String("private", "", "Path to the private key file")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'create' or 'load' subcommands")
		os.Exit(1)
	}

	var device tpmutil.Device
	if *useTPM {
		device = tpmutil.LINUX
	} else {
		device = tpmutil.SWTPM
	}

	switch subcmd := os.Args[1]; subcmd {
	case "create", "load":
		tpm, err := tpmutil.OpenTPM(device)
		if err != nil {
			log.Fatalf("can't open tpm: %v", err)
		}
		defer tpm.Close()

		if subcmd == "create" {
			createCmd.Parse(os.Args[2:])
			if err := createOrdinaryKey(tpm, *createOutDir); err != nil {
				fmt.Printf("Error creating ordinary key: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Ordinary key created successfully 🚀")
		} else {
			loadCmd.Parse(os.Args[2:])
			if *loadPub == "" || *loadPrivate == "" {
				fmt.Println("Both --public and --private flags are required for 'load'")
				os.Exit(1)
			}
			if err := loadOrdinaryKey(tpm, *loadPub, *loadPrivate); err != nil {
				fmt.Printf("Error loading ordinary key: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Ordinary key loaded successfully 🚀")
		}
	case "cleanup":
		if err := os.RemoveAll(tpmutil.SWTPM_ROOT_STATE); err != nil {
			fmt.Printf("Error cleaning state: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("State cleaned successfully 🚀")
	default:
		fmt.Println("Unknown subcommand. Expected 'create' or 'load'")
		os.Exit(1)
	}
}

func createOrdinaryKey(tpm transport.TPM, outDir string) error {
	if outDir == "" {
		if dir, err := os.Getwd(); err == nil {
			outDir = dir
		}
	}
	primary, err, primaryCloser := createPrimary(tpm, ECCStorageParentTemplate)
	if err != nil {
		return fmt.Errorf("failed to create primary key failed: %w", err)
	}
	defer primaryCloser()

	createRsp, err := tpm2.Create{
		ParentHandle: tpm2.NamedHandle{
			Name:   primary.Name,
			Handle: primary.ObjectHandle,
		},
		InPublic: ECCSignerTemplate,
	}.Execute(tpm)
	if err != nil {
		return fmt.Errorf("failed to create ordinary key: %w", err)
	}

	// Save the TPM2B_PUBLIC
	if err := os.WriteFile(path.Join(outDir, "tpmkey.pub"), tpm2.Marshal(createRsp.OutPublic), 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}
	// Save the TPM2B_PRIVATE
	if err := os.WriteFile(path.Join(outDir, "tpmkey.priv"), tpm2.Marshal(createRsp.OutPrivate), 0644); err != nil {
		return fmt.Errorf("failed to write private blob: %w", err)
	}
	return nil
}

func loadOrdinaryKey(tpm transport.TPM, pubPath, privPath string) error {
	pub, priv, err := loadTPMBlob(pubPath, privPath)
	if err != nil {
		return fmt.Errorf("failed to get public and private keys: %w", err)
	}

	primary, err, primaryCloser := createPrimary(tpm, ECCStorageParentTemplate)
	if err != nil {
		return fmt.Errorf("failed to create primary key failed: %w", err)
	}
	defer primaryCloser()

	_, err = tpm2.Load{
		ParentHandle: tpm2.NamedHandle{
			Name:   primary.Name,
			Handle: primary.ObjectHandle,
		},
		InPublic:  *pub,
		InPrivate: *priv,
	}.Execute(tpm)
	if err != nil {
		return fmt.Errorf("failed to load ordinary key: %w", err)
	}
	return nil
}

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
