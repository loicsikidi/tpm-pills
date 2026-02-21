package tpmutil

import "github.com/loicsikidi/go-tpm-kit/tpmutil"

type Handle = tpmutil.Handle
type HandleCloser = tpmutil.HandleCloser

var (
	MustGenerateRnd       = tpmutil.MustGenerateRnd
	Persist               = tpmutil.Persist
	GetPersistedKeyHandle = tpmutil.GetPersistedKeyHandle
	NewHandle             = tpmutil.NewHandle
	ToAuthHandle          = tpmutil.ToAuthHandle
)

type PersistConfig = tpmutil.PersistConfig
type GetPersistedKeyHandleConfig = tpmutil.GetPersistedKeyHandleConfig
