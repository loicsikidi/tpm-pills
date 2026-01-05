# Pill #5

## Goal

The goal of this example is to show how to:

1. decrypt an encrypted blob using `TPM2_RSA_Decrypt`
1. sign a message using a non restricted signing key
1. sign a message using a restricted signing key

[`rsa_encryption_test`](./rsa_encryption_test.go) on its part demonstrates two concepts described in the pill:

1. a *Primary key* is reproducible
1. only *Storage keys* can be used to create an *Ordinary key*

### Prerequisites

This example requires `swtpm` installed on your running system. Read [pill #2](https://tpmpills.com/02-install-tooling.html) to learn how to obtain a proper environment.

## Run the examples

> [!TIP]
> Examples use a Software TPM (i.e swtpm).
> If you want to rely on a real TPM, add the `--use-real-tpm` flag to the command.

### Encrypt/Decrypt a blob

```bash
# Create the decryption key
# Note: the key will be stored in the current directory with the name `key.tpm` and `public.pem`
go run github.com/loicsikidi/tpm-pills/examples/05-pill create --type decrypt

# Encrypt a blob using the public key
go run github.com/loicsikidi/tpm-pills/examples/05-pill encrypt --key ./public.pem --message 'Hello TPM Pills!' --output ./blob.enc

# Alternatively, you can use the `openssl` command to encrypt the blob
openssl pkeyutl -encrypt -in <(echo -n 'Hello TPM Pills!') -out ./blob.enc -pubin -inkey public.pem -pkeyopt rsa_padding_mode:oaep -pkeyopt rsa_oaep_md:sha256

# Decrypt the blob using the private key held in the TPM
go run github.com/loicsikidi/tpm-pills/examples/05-pill decrypt --key ./key.tpm --in ./blob.enc

# Clean up
# Note:
# 1. the command will remove swtpm state
# 2. the command is optional if --use-real-tpm flag is set
go run github.com/loicsikidi/tpm-pills/examples/05-pill cleanup

# remove created files
rm -f ./key.tpm ./public.pem ./blob.enc
```

### Sign/Verify a message with a non restricted signing key

```bash
# Create the non-restricted signing key
# Note: the key will be stored in the current directory with the name `key.tpm` and `public.pem`
go run github.com/loicsikidi/tpm-pills/examples/05-pill create --type signer

# Sign a message using the private key held in the TPM
go run github.com/loicsikidi/tpm-pills/examples/05-pill sign --key ./key.tpm --message 'Hello TPM Pills!' --output ./message.sig
# output: Signature saved to ./message.sig 🚀

# Verify the signature using the public key
go run github.com/loicsikidi/tpm-pills/examples/05-pill verify --key ./public.pem --signature ./message.sig --message 'Hello TPM Pills!'
# output: Signature verified successfully 🚀

# Alternatively, you can use the `openssl` command to verify the signature
openssl dgst -sha256 -verify ./public.pem -signature ./message.sig <(echo -n 'Hello TPM Pills!')
# output: Verified OK

# Clean up
# Note:
# 1. the command will remove swtpm state
# 2. the command is optional if --use-real-tpm flag is set
go run github.com/loicsikidi/tpm-pills/examples/05-pill cleanup

# remove created files
rm -f ./key.tpm ./public.pem ./message.sig
```

### Sign/Verify a message with a restricted signing key

```bash
# Create the restricted signing key
# Note: the key will be stored in the current directory with the name `key.tpm` and `public.pem`
go run github.com/loicsikidi/tpm-pills/examples/05-pill create --type restrictedSigner

# Sign a message using the private key held in the TPM
go run github.com/loicsikidi/tpm-pills/examples/05-pill sign --key ./key.tpm --message 'Hello TPM Pills!' --output ./message.sig
# output: Signature saved to ./message.sig 🚀

# Verify the signature using the public key
go run github.com/loicsikidi/tpm-pills/examples/05-pill verify --key ./public.pem --signature ./message.sig --message 'Hello TPM Pills!'
# output: Signature verified successfully 🚀

# Alternatively, you can use the `openssl` command to verify the signature
openssl dgst -sha256 -verify ./public.pem -signature ./message.sig <(echo -n 'Hello TPM Pills!')
# output: Verified OK

# Clean up
# Note:
# 1. the command will remove swtpm state
# 2. the command is optional if --use-real-tpm flag is set
go run github.com/loicsikidi/tpm-pills/examples/05-pill cleanup

# remove created files
rm -f ./key.tpm ./public.pem ./message.sig
```

## Run tests

```bash
# Run the tests
go test -v github.com/loicsikidi/tpm-pills/examples/05-pill

## Run all tests 
go test -v github.com/loicsikidi/tpm-pills/examples/05-pill -args --tag=all-tests
```
