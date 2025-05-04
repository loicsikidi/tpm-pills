# Pill #4

## Goal

The goal of this example is to show how to:

1. create an ordinary key (ECC NIST P256) and to store its content in the filesystem
1. load the key into the TPM

[`create_key_test`](./create_key_test.go) on its part demonstrates two concepts described in the pill:

1. a *Primary key* is reproducible
1. only *Storage keys* can be used to create an *Ordinary key*

### Prerequisites

This example requires `swtpm` installed on your running system. Read [pill #2](https://tpmpills.com/02-install-tooling.html) to learn how to obtain a proper environment.

## Run the example

```bash
# Create the key
# Note: the key will be stored in the current directory with the name `tpmkey.pub` and `tpmkey.priv`
go run github.com/loicsikidi/tpm-pills/examples/04-pill create

# Load the key
go run github.com/loicsikidi/tpm-pills/examples/04-pill load --public ./tpmkey.pub --private ./tpmkey.priv

# Clean up
# Note: the command wll remove swtpm state
go run github.com/loicsikidi/tpm-pills/examples/04-pill cleanup
```

## Run tests

```bash
# Run the tests
go test github.com/loicsikidi/tpm-pills/examples/04-pill
```




