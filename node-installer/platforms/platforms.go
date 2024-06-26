// Copyright 2024 Edgeless Systems GmbH
// SPDX-License-Identifier: AGPL-3.0-only

// The platforms package provides a constant interface to the different deployment platforms
// of Contrast.
package platforms

import "fmt"

// Platform is a type that represents a deployment platform of Contrast.
type Platform int

const (
	// Unknown is the default value for the platform type.
	Unknown Platform = iota
	// AKSCloudHypervisorSNP represents a deployment with Cloud-Hypervisor on SEV-SNP AKS.
	AKSCloudHypervisorSNP
	// K3sQEMUTDX represents a deployment with QEMU on bare-metal TDX K3s.
	K3sQEMUTDX
	// RKE2QEMUTDX represents a deployment with QEMU on bare-metal TDX RKE2.
	RKE2QEMUTDX
)

// String returns the string representation of the Platform type.
func (p Platform) String() string {
	switch p {
	case AKSCloudHypervisorSNP:
		return "AKS-CLH-SNP"
	case K3sQEMUTDX:
		return "K3s-QEMU-TDX"
	case RKE2QEMUTDX:
		return "RKE2-QEMU-TDX"
	default:
		return "Unknown"
	}
}

// FromString returns the Platform type corresponding to the given string.
func FromString(s string) (Platform, error) {
	switch s {
	case "AKS-CLH-SNP":
		return AKSCloudHypervisorSNP, nil
	case "K3s-QEMU-TDX":
		return K3sQEMUTDX, nil
	case "RKE2-QEMU-TDX":
		return RKE2QEMUTDX, nil
	default:
		return Unknown, fmt.Errorf("unknown platform: %s", s)
	}
}
