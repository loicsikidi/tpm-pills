//go:build !windows

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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
	loadPub := loadCmd.String("public", "", "Path to TPM public file")
	loadPrivate := loadCmd.String("private", "", "Path to TPM private file")

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
			if err := tpmutil.CreateOrdinaryKey(tpm,
				*createOutDir,
				tpmutil.ECCP256StorageParentTemplate,
				tpmutil.ECCP256SignerTemplate,
				/* createPublicKey */ true); err != nil {
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
			if _, closer, err := tpmutil.LoadOrdinaryKey(tpm,
				tpmutil.ECCP256StorageParentTemplate,
				*loadPub,
				*loadPrivate); err != nil {
				fmt.Printf("Error loading ordinary key: %v\n", err)
				os.Exit(1)
			} else {
				closer() // flush context
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
		fmt.Println("Unknown subcommand. Expected 'create', 'load' or 'cleanup'")
		os.Exit(1)
	}
}
