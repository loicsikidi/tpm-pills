package options

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/loicsikidi/tpm-pills/internal/utils"
)

type KeyType int

const (
	UnspecifiedKeyType KeyType = iota
	Decrypt
	Signer
	RestrictedSigner
)

const (
	defaultKeyFileName       = "key.tpm"
	defaultSealedFileName    = "sealed_key.tpm"
	defaultEncryptedFileName = "blob.enc"
	defaultSignedFileName    = "message.sig"
	defaultHandleStr         = "0x81000010"
)

var validKeyTypes = []KeyType{
	UnspecifiedKeyType,
	Decrypt,
	Signer,
	RestrictedSigner,
}

func (k KeyType) String() string {
	switch k {
	case UnspecifiedKeyType:
		return "unspecified"
	case Decrypt:
		return "decrypt"
	case Signer:
		return "signer"
	case RestrictedSigner:
		return "restricted-signer"
	default:
		return fmt.Sprintf("unknown(%d)", k)
	}
}

func (k KeyType) Check() error {
	switch k {
	case UnspecifiedKeyType, Decrypt, Signer, RestrictedSigner:
		return nil
	default:
		return fmt.Errorf("invalid KeyType: %d", k)
	}
}

func FromStringToKeyType(s string) KeyType {
	for _, kt := range validKeyTypes {
		if strings.EqualFold(kt.String(), s) {
			return kt
		}
	}
	return -1
}

type CreateKeyOpts struct {
	OutputDir string
	KeyType   string
	kty       KeyType
}

func (o *CreateKeyOpts) CheckAndSetDefaults() error {
	if o.OutputDir == "" {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("invalid input: failed to fallback to a default 'OutputDir': %w", err)
		}
		o.OutputDir = dir
	}
	o.kty = FromStringToKeyType(o.KeyType)
	if err := o.kty.Check(); err != nil {
		return fmt.Errorf("invalid input: invalid KeyType: %w", err)
	}
	if o.kty == UnspecifiedKeyType {
		o.kty = Signer
	}
	return nil
}

func (o *CreateKeyOpts) GetKeyType() KeyType {
	return o.kty
}

type EncryptOpts struct {
	PublicKeyPath  string
	Message        string
	OutputFilePath string
}

func (o *EncryptOpts) CheckAndSetDefaults() error {
	if o.PublicKeyPath == "" {
		return fmt.Errorf("invalid input: PublicKeyPath is required")
	}
	if !utils.FileExists(o.PublicKeyPath) {
		return fmt.Errorf("invalid input: PublicKeyPath does not exist")
	}
	if len(o.Message) == 0 {
		return fmt.Errorf("invalid input: Message is required")
	}
	if o.OutputFilePath == "" {
		dir, err := utils.FallbackDir()
		if err != nil {
			return err
		}
		o.OutputFilePath = filepath.Join(dir, defaultEncryptedFileName)
	}
	return nil
}

type SymEncryptOpts struct {
	KeyBlobPath    string
	Message        string
	OutputFilePath string
}

func (o *SymEncryptOpts) CheckAndSetDefaults() error {
	dir, err := utils.FallbackDir()
	if err != nil {
		return err
	}
	if o.KeyBlobPath == "" {
		o.KeyBlobPath = filepath.Join(dir, defaultKeyFileName)
	}
	if !utils.FileExists(o.KeyBlobPath) {
		return fmt.Errorf("invalid input: KeyBlobPath does not exist")
	}
	if len(o.Message) == 0 {
		return fmt.Errorf("invalid input: Message is required")
	}
	if o.OutputFilePath == "" {
		o.OutputFilePath = filepath.Join(dir, defaultEncryptedFileName)
	}
	if !utils.DirExists(filepath.Dir(o.OutputFilePath)) {
		return fmt.Errorf("invalid input: OutputFilePath parent directory does not exist")
	}
	return nil
}

type DecryptOpts struct {
	InputFilePath string
	KeyBlobPath   string
}

func (o *DecryptOpts) CheckAndSetDefaults() error {
	dir, err := utils.FallbackDir()
	if err != nil {
		return err
	}
	if o.InputFilePath == "" {
		o.InputFilePath = filepath.Join(dir, defaultEncryptedFileName)
	}
	if !utils.FileExists(o.InputFilePath) {
		return fmt.Errorf("invalid input: InputFilePath does not exist")
	}
	if o.KeyBlobPath == "" {
		o.KeyBlobPath = filepath.Join(dir, defaultKeyFileName)
	}
	if !utils.FileExists(o.KeyBlobPath) {
		return fmt.Errorf("invalid input: KeyBlobPath does not exist")
	}
	return nil
}

type AsymDecryptOpts struct {
	InputFilePath string
	KeyBlobPath   string
}

func (o *AsymDecryptOpts) CheckAndSetDefaults() error {
	dir, err := utils.FallbackDir()
	if err != nil {
		return err
	}
	if o.InputFilePath == "" {
		o.InputFilePath = filepath.Join(dir, defaultEncryptedFileName)
	}
	if !utils.FileExists(o.InputFilePath) {
		return fmt.Errorf("invalid input: InputFilePath does not exist")
	}
	if o.KeyBlobPath == "" {
		o.KeyBlobPath = filepath.Join(dir, defaultKeyFileName)
	}
	if !utils.FileExists(o.KeyBlobPath) {
		return fmt.Errorf("invalid input: KeyBlobPath does not exist")
	}
	return nil
}

type SignOpts struct {
	KeyBlobPath    string
	Message        string
	OutputFilePath string
}

func (o *SignOpts) CheckAndSetDefaults() error {
	dir, err := utils.FallbackDir()
	if err != nil {
		return err
	}
	if o.KeyBlobPath == "" {
		o.KeyBlobPath = filepath.Join(dir, defaultKeyFileName)
	}
	if !utils.FileExists(o.KeyBlobPath) {
		return fmt.Errorf("invalid input: KeyBlobPath does not exist")
	}
	if len(o.Message) == 0 {
		return fmt.Errorf("invalid input: Message is required")
	}
	if o.OutputFilePath == "" {
		o.OutputFilePath = filepath.Join(dir, defaultSignedFileName)
	}
	if !utils.DirExists(filepath.Dir(o.OutputFilePath)) {
		return fmt.Errorf("invalid input: OutputFilePath parent directory does not exist")
	}
	return nil
}

type VerifyOpts struct {
	PublicKeyPath string
	Message       string
	SignaturePath string
}

func (o *VerifyOpts) CheckAndSetDefaults() error {
	if o.PublicKeyPath == "" {
		return fmt.Errorf("invalid input: PublicKeyPath is required")
	}
	if !utils.FileExists(o.PublicKeyPath) {
		return fmt.Errorf("invalid input: PublicKeyPath does not exist")
	}
	if len(o.Message) == 0 {
		return fmt.Errorf("invalid input: Message is required")
	}
	if o.SignaturePath == "" {
		dir, err := utils.FallbackDir()
		if err != nil {
			return err
		}
		o.SignaturePath = filepath.Join(dir, defaultSignedFileName)
	}
	if !utils.FileExists(o.SignaturePath) {
		return fmt.Errorf("invalid input: SignaturePath does not exist")
	}
	return nil
}

type SealOpts struct {
	Message        string
	OutputFilePath string
}

func (o *SealOpts) CheckAndSetDefaults() error {
	if len(o.Message) == 0 {
		return fmt.Errorf("invalid input: Message is required")
	}
	if o.OutputFilePath == "" {
		dir, err := utils.FallbackDir()
		if err != nil {
			return err
		}
		o.OutputFilePath = filepath.Join(dir, defaultSealedFileName)
	}
	if !utils.DirExists(filepath.Dir(o.OutputFilePath)) {
		return fmt.Errorf("invalid input: OutputFilePath parent directory does not exist")
	}
	return nil
}

type UnsealOpts struct {
	InputFilePath string
}

func (o *UnsealOpts) CheckAndSetDefaults() error {
	if o.InputFilePath == "" {
		dir, err := utils.FallbackDir()
		if err != nil {
			return err
		}
		o.InputFilePath = filepath.Join(dir, defaultSealedFileName)
	}
	if !utils.FileExists(o.InputFilePath) {
		return fmt.Errorf("invalid input: InputFilePath does not exist")
	}
	return nil
}

type HMACOpts struct {
	Data string
}

func (o *HMACOpts) CheckAndSetDefaults() error {
	if len(o.Data) == 0 {
		return fmt.Errorf("invalid input: Data is required")
	}
	return nil
}

type PersistOpts struct {
	Handle    string
	OutputDir string
}

func (o *PersistOpts) CheckAndSetDefaults() error {
	if o.Handle == "" {
		o.Handle = defaultHandleStr
	}
	if o.OutputDir == "" {
		dir, err := utils.FallbackDir()
		if err != nil {
			return err
		}
		o.OutputDir = dir
	}
	return nil
}

type ReadPersistedOpts struct {
	Handle        string
	PublicKeyPath string
}

func (o *ReadPersistedOpts) CheckAndSetDefaults() error {
	if o.Handle == "" {
		o.Handle = defaultHandleStr
	}
	if o.PublicKeyPath == "" {
		return fmt.Errorf("invalid input: PublicKeyPath is required")
	}
	if !utils.FileExists(o.PublicKeyPath) {
		return fmt.Errorf("invalid input: %s does not exist", o.PublicKeyPath)
	}
	return nil
}

type UnpersistOpts struct {
	Handle string
}

func (o *UnpersistOpts) CheckAndSetDefaults() error {
	if o.Handle == "" {
		o.Handle = defaultHandleStr
	}
	return nil
}
