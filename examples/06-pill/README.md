# Pill #6

## Goal

The goal of this example is to show how to:

1. encrypt/decrypt data using symmetric keys with `TPM2_EncryptDecrypt2`
1. seal/unseal data using `TPM2_Create` and `TPM2_Unseal`
1. compute HMAC signatures using `TPM2_HMAC`

[`concepts_test`](./concepts_test.go) on its part demonstrates some concepts described in the pill:

1. maximum seal size is 128 bytes
1. HMAC output sizes vary by hash algorithm

### Prerequisites

This example requires `swtpm` installed on your running system. Read [pill #2](https://tpmpills.com/02-install-tooling.html) to learn how to obtain a proper environment.

## Run the examples

> [!TIP]
> Examples use a Software TPM (i.e swtpm).
> If you want to rely on a real TPM, add the `--use-real-tpm` flag to the command.

### Encrypt/Decrypt with symmetric keys

```bash
# Create the symmetric key
# Note: the key will be stored in the current directory with the name `key.tpm`
go run github.com/loicsikidi/tpm-pills/examples/06-pill create

# Encrypt a message
go run github.com/loicsikidi/tpm-pills/examples/06-pill encrypt --message "Hello TPM Pills!" --output ./blob.enc

# Decrypt the message
go run github.com/loicsikidi/tpm-pills/examples/06-pill decrypt --key ./key.tpm --in ./blob.enc

# Clean up
go run github.com/loicsikidi/tpm-pills/examples/06-pill cleanup
rm -f ./key.tpm ./blob.enc
```

### Seal/Unseal data

```bash
# Seal a message
go run github.com/loicsikidi/tpm-pills/examples/06-pill seal --message "important secret" --output ./sealed_key.tpm

# Unseal the message
go run github.com/loicsikidi/tpm-pills/examples/06-pill unseal --in ./sealed_key.tpm

# Clean up
go run github.com/loicsikidi/tpm-pills/examples/06-pill cleanup
rm -f ./sealed_key.tpm
```

### Compute HMAC

```bash
# Compute HMAC for data
go run github.com/loicsikidi/tpm-pills/examples/06-pill hmac --data "secret"
# output: HMAC result: "$HEX VALUE" 🚀

# Verify deterministic output
go run github.com/loicsikidi/tpm-pills/examples/06-pill hmac --data "secret"
# output: HMAC result: "$HEX VALUE" 🚀
# Clean up
go run github.com/loicsikidi/tpm-pills/examples/06-pill cleanup
```

## Run tests

```bash
# Run the tests
go test -v github.com/loicsikidi/tpm-pills/examples/06-pill
```
