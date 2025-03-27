package main

import (
	"encoding/binary"
	"fmt"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpm2/transport"
)

func getTpmVersion(tpm transport.TPMCloser) ([]byte, error) {
	getCapCmd := tpm2.GetCapability{
		Capability:    tpm2.TPMCapTPMProperties,
		Property:      uint32(tpm2.TPMPTFamilyIndicator),
		PropertyCount: 1,
	}
	getCapRsp, err := getCapCmd.Execute(tpm)
	if err != nil {
		return nil, fmt.Errorf("cmd.GetCapability() failed: %w", err)
	}

	props, err := getCapRsp.CapabilityData.Data.TPMProperties()
	if err != nil {
		return nil, fmt.Errorf("tpm2.TPMUCapabilities.TPMProperties() failed: %w", err)
	}

	// value is stored on 4 octet
	version := make([]byte, 4)
	binary.BigEndian.PutUint32(version, props.TPMProperty[0].Value)
	return version, nil
}
