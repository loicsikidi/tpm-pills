# Pill #4

## Goal

The goal of this example is to show how to:

1. create an ordinary key (ECC NIST P256) and to store its content in the filesystem
1. load the key into the TPM

[`concepts_test.go`](./concepts_test.go) on its part demonstrates two concepts described in the pill:

1. a *Primary key* is reproducible
1. only *Storage keys* can be used to create an *Ordinary key*

### Prerequisites

This example requires `swtpm` installed on your running system. Read [pill #2](https://tpmpills.com/02-install-tooling.html) to learn how to obtain a proper environment.

## Run the example

> [!TIP]
> Examples use a Software TPM (i.e swtpm).
> If you want to rely on a real TPM, add the `--use-real-tpm` flag to the command.

```bash
# Create the key
# Note: the key will be stored in the current directory with the name `key.tpm` and `public.pem`
go run github.com/loicsikidi/tpm-pills/examples/04-pill create

# Load the key
go run github.com/loicsikidi/tpm-pills/examples/04-pill load --key ./key.tpm

# Clean up
# Note:
# 1. the command will remove swtpm state
# 2. the command is optional if --use-real-tpm flag is set
go run github.com/loicsikidi/tpm-pills/examples/04-pill cleanup

# remove created files
rm -f ./key.tpm
```

## Run tests

```bash
# Run the tests
go test -v github.com/loicsikidi/tpm-pills/examples/04-pill
```
