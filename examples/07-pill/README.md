# Pill #7

## Goal

The goal of this example is to show how to:

1. persist a key at a given TPM handle using `TPM2_EvictControl`
1. read a persisted key and verify it matches a known public key
1. unpersist (evict) a key from persistent storage using `TPM2_EvictControl`

### Prerequisites

This example requires `swtpm` installed on your running system. Read [pill #2](https://tpmpills.com/02-install-tooling.html) to learn how to obtain a proper environment.

## Run the examples

> [!TIP]
> Examples use a Software TPM (i.e swtpm).
> If you want to rely on a real TPM, add the `--use-real-tpm` flag to the command.

### Persist a key

```bash
# Create and persist an ECC key at the default handle (0x81000010)
# The command outputs the public key in PEM format
go run github.com/loicsikidi/tpm-pills/examples/07-pill persist

# Persist a key at a custom handle
go run github.com/loicsikidi/tpm-pills/examples/07-pill persist --handle 0x81000020
```

### Read and verify a persisted key

```bash
# Verify that the persisted key matches a known public key
go run github.com/loicsikidi/tpm-pills/examples/07-pill read --pubkey ./public.pem

# Read from a custom handle
go run github.com/loicsikidi/tpm-pills/examples/07-pill read --handle 0x81000020 --pubkey ./public.pem
```

### Unpersist a key

```bash
# Remove the persisted key at the default handle
go run github.com/loicsikidi/tpm-pills/examples/07-pill unpersist

# Remove a key at a custom handle
go run github.com/loicsikidi/tpm-pills/examples/07-pill unpersist --handle 0x81000020

# Clean up swtpm state
go run github.com/loicsikidi/tpm-pills/examples/07-pill cleanup
```

## Run tests

```bash
# Run the tests
go test -v github.com/loicsikidi/tpm-pills/examples/07-pill
```
