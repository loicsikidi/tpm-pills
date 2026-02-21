//go:build !windows

package main

import (
	"crypto"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport"
	"github.com/loicsikidi/go-tpm-kit/tpmcrypto"
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
	persistOpts := &options.PersistOpts{}
	readOpts := &options.ReadPersistedOpts{}
	unpersistOpts := &options.UnpersistOpts{}

	persistCmd := flag.NewFlagSet("persist", flag.ExitOnError)
	persistCmd.StringVar(&persistOpts.Handle, "handle", "", "Target persistent handle (default: 0x81000010)")
	persistCmd.StringVar(&persistOpts.OutputDir, "out", "", "Output directory for the created key")
	persistCmd.BoolVar(&useTPM, "use-real-tpm", false, "Use real TPM instead of swtpm")

	readCmd := flag.NewFlagSet("read", flag.ExitOnError)
	readCmd.StringVar(&readOpts.Handle, "handle", "", "Target persistent handle (default: 0x81000010)")
	readCmd.StringVar(&readOpts.PublicKeyPath, "pubkey", "", "Path to the public key file")
	readCmd.BoolVar(&useTPM, "use-real-tpm", false, "Use real TPM instead of swtpm")

	unpersistCmd := flag.NewFlagSet("unpersist", flag.ExitOnError)
	unpersistCmd.StringVar(&unpersistOpts.Handle, "handle", "", "Target persistent handle (default: 0x81000010)")
	unpersistCmd.BoolVar(&useTPM, "use-real-tpm", false, "Use real TPM instead of swtpm")

	if len(os.Args) < 2 {
		flag.Usage()
		return fmt.Errorf("missing subcommand")
	}

	switch subcmd := os.Args[1]; subcmd {
	case "persist", "read", "unpersist":
		switch subcmd {
		case "persist":
			persistCmd.Parse(os.Args[2:])
		case "read":
			readCmd.Parse(os.Args[2:])
		case "unpersist":
			unpersistCmd.Parse(os.Args[2:])
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

		if subcmd == "persist" {
			if err := persistCommand(tpm, persistOpts); err != nil {
				return fmt.Errorf("error persisting key: %w", err)
			}
			fmt.Printf("Key persisted at handle %s\n", persistOpts.Handle)
			fmt.Printf("Public key saved to %s 🚀\n", filepath.Join(persistOpts.OutputDir, "public.pem"))
		}
		if subcmd == "read" {
			if err := readCommand(tpm, readOpts); err != nil {
				return fmt.Errorf("error reading persisted key: %w", err)
			}
			fmt.Printf("Persisted key at handle %s matches the provided public key ✅\n", readOpts.Handle)
		}
		if subcmd == "unpersist" {
			if err := unpersistCommand(tpm, unpersistOpts); err != nil {
				return fmt.Errorf("error unpersisting key: %w", err)
			}
			fmt.Printf("Key at handle %s has been removed\n", unpersistOpts.Handle)
		}
	case "cleanup":
		if err := os.RemoveAll(tpmutil.SWTPM_ROOT_STATE); err != nil {
			return fmt.Errorf("error cleaning state: %w", err)
		}
		fmt.Println("State cleaned successfully")
	default:
		return fmt.Errorf("unknown subcommand %q. Expected 'persist', 'read', 'unpersist' or 'cleanup'", subcmd)
	}
	return nil
}

// persistCommand creates an ECC ordinary key using [tpmutil.CreateKey], loads it
// back into the TPM, persists it at the given handle, and returns the public key
// in PEM format.
func persistCommand(tpm transport.TPM, opts *options.PersistOpts) error {
	if err := opts.CheckAndSetDefaults(); err != nil {
		return err
	}

	handle, err := parseHandle(opts.Handle)
	if err != nil {
		return err
	}

	// 1. Create an ECC ordinary key (saves key.tpm + public.pem to OutputDir)
	if err := tpmutil.CreateKey(tpm, tpmutil.CreateKeyConfig{
		OutDir:           opts.OutputDir,
		ParentTemplate:   tpmutil.ECCSRKTemplate,
		OrdinaryTemplate: tpmutil.ECCSignerTemplate,
		CreatePublicKey:  true,
	}); err != nil {
		return fmt.Errorf("failed to create key: %w", err)
	}

	// 2. Load the created key back into the TPM
	keyBlobPath := filepath.Join(opts.OutputDir, "key.tpm")
	keyHandle, err := tpmutil.LoadKey(tpm, tpmutil.LoadKeyConfig{
		ParentTemplate: tpmutil.ECCSRKTemplate,
		KeyBlobPath:    keyBlobPath,
	})
	if err != nil {
		return fmt.Errorf("failed to load key: %w", err)
	}
	defer keyHandle.Close()

	// 3. Persist the key at the target handle
	if _, err = tpmutil.Persist(tpm, tpmutil.PersistConfig{
		TransientHandle:  keyHandle,
		PersistentHandle: tpmutil.NewHandle(handle),
	}); err != nil {
		return fmt.Errorf("failed to persist key: %w", err)
	}

	// 4. Remove the key blob file (because not necessary for the rest)
	if err := os.Remove(keyBlobPath); err != nil {
		return fmt.Errorf("failed to remove key blob: %w", err)
	}
	return nil
}

// readCommand loads the persisted key handle, extracts its public key,
// and compares it with the public key file provided via --pubkey.
func readCommand(tpm transport.TPM, opts *options.ReadPersistedOpts) error {
	if err := opts.CheckAndSetDefaults(); err != nil {
		return err
	}

	handle, err := parseHandle(opts.Handle)
	if err != nil {
		return err
	}

	// 1. Read the persisted handle
	persistedHandle, err := tpmutil.GetPersistedKeyHandle(tpm, tpmutil.GetPersistedKeyHandleConfig{
		Handle: tpmutil.NewHandle(handle),
	})
	if err != nil {
		return fmt.Errorf("failed to get persisted key handle: %w", err)
	}

	// 2. Extract the public key from the persisted handle
	persistedPub, err := tpmcrypto.PublicKey(persistedHandle.Public())
	if err != nil {
		return fmt.Errorf("failed to extract public key from persisted handle: %w", err)
	}

	// 3. Read and parse the public key from the file
	filePub, err := pemutil.Read(opts.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key file: %w", err)
	}

	// 4. Compare the public keys
	filePubKey, ok := filePub.(crypto.PublicKey)
	if !ok {
		return fmt.Errorf("file key is not a valid public key")
	}

	pub, ok := persistedPub.(interface{ Equal(x crypto.PublicKey) bool })
	if !ok {
		return fmt.Errorf("invalid public key: Equal is not implemented")
	}

	if !pub.Equal(filePubKey) {
		return fmt.Errorf("public keys do not match")
	}

	return nil
}

// unpersistCommand removes a persisted key from the given handle using [tpm2.EvictControl].
func unpersistCommand(tpm transport.TPM, opts *options.UnpersistOpts) error {
	if err := opts.CheckAndSetDefaults(); err != nil {
		return err
	}

	handle, err := parseHandle(opts.Handle)
	if err != nil {
		return err
	}

	keyHandle, err := tpmutil.GetPersistedKeyHandle(tpm, tpmutil.GetPersistedKeyHandleConfig{
		Handle: tpmutil.NewHandle(handle),
	})
	if err != nil {
		return fmt.Errorf("failed to get persisted key handle: %w", err)
	}

	// Remove the key from persistent storage
	_, err = tpm2.EvictControl{
		Auth:             tpmutil.ToAuthHandle(tpmutil.NewHandle(tpm2.TPMRHOwner)),
		ObjectHandle:     keyHandle,
		PersistentHandle: handle,
	}.Execute(tpm)
	if err != nil {
		return fmt.Errorf("failed to evict key at handle 0x%x: %w", handle, err)
	}

	return nil
}

// parseHandle parses a hex string (e.g. "0x81000010") into a [tpm2.TPMHandle].
func parseHandle(s string) (tpm2.TPMHandle, error) {
	v, err := strconv.ParseUint(s, 0, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid handle %q: %w", s, err)
	}
	return tpm2.TPMHandle(v), nil
}
