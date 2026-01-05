//go:build !windows

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/google/go-tpm/tpm2/transport"
	"github.com/loicsikidi/tpm-pills/internal/options"
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
		KeyType: options.Signer.String(),
	}
	loadOpts := &LoadKeyOpts{}

	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	loadCmd := flag.NewFlagSet("load", flag.ExitOnError)

	// Define flags for the create subcommand
	createCmd.StringVar(&createOpts.OutputDir, "out", "", "Output directory for the created key")
	createCmd.BoolVar(&useTPM, "use-real-tpm", false, "Use real TPM instead of swtpm")

	// Define flags for the load subcommand
	loadCmd.StringVar(&loadOpts.KeyBlobPath, "key", "", "Path to TPM key blob file")
	loadCmd.BoolVar(&useTPM, "use-real-tpm", false, "Use real TPM instead of swtpm")

	if len(os.Args) < 2 {
		flag.Usage()
		return fmt.Errorf("missing subcommand")
	}

	switch subcmd := os.Args[1]; subcmd {
	case "create", "load":
		switch subcmd {
		case "create":
			createCmd.Parse(os.Args[2:])
		case "load":
			loadCmd.Parse(os.Args[2:])
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
		} else {
			if err := loadCommand(tpm, loadOpts); err != nil {
				return fmt.Errorf("error loading key: %w", err)
			}
			fmt.Println("Ordinary key loaded successfully 🚀")
		}
	case "cleanup":
		if err := os.RemoveAll(tpmutil.SWTPM_ROOT_STATE); err != nil {
			return fmt.Errorf("error cleaning state: %w", err)
		}
		fmt.Println("State cleaned successfully 🚀")
	default:
		return fmt.Errorf("unknown subcommand %q. Expected 'create', 'load' or 'cleanup'", subcmd)
	}
	return nil
}

type LoadKeyOpts struct {
	KeyBlobPath string
}

func (o *LoadKeyOpts) CheckAndSetDefaults() error {
	if o.KeyBlobPath == "" {
		return fmt.Errorf("invalid input: KeyBlobPath is required")
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
		OrdinaryTemplate: tpmutil.ECCSignerTemplate,
		CreatePublicKey:  false,
	})
}

func loadCommand(tpm transport.TPM, opts *LoadKeyOpts) error {
	if err := opts.CheckAndSetDefaults(); err != nil {
		return err
	}

	keyHandle, err := tpmutil.LoadKey(tpm, tpmutil.LoadKeyConfig{
		ParentTemplate: tpmutil.ECCSRKTemplate,
		KeyBlobPath:    opts.KeyBlobPath,
	})
	if err != nil {
		return fmt.Errorf("failed to load key: %w", err)
	}
	defer keyHandle.Close()

	return nil
}
