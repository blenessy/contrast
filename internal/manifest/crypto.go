// Copyright 2024 Edgeless Systems GmbH
// SPDX-License-Identifier: AGPL-3.0-only

package manifest

import (
"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
"encoding/pem"
	"fmt"

	"github.com/edgelesssys/contrast/internal/userapi"
)

// NewSeedShareOwnerPrivateKey creates and PEM-encodes a new seed share private key.
func NewSeedShareOwnerPrivateKey() ([]byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("generating private key: %w", err)
	}
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateKeyBytes}), nil
}

// MarshalSeedShareOwnerKey converts a public key into the format for userapi.SetManifestRequest.
func MarshalSeedShareOwnerKey(pubKey *rsa.PublicKey) HexString {
	return NewHexString(x509.MarshalPKCS1PublicKey(pubKey))
}

// ParseSeedShareOwnerKey reads a public key embedded in a userapi.SetManifestRequest.
func ParseSeedShareOwnerKey(pubKeyHex HexString) (*rsa.PublicKey, error) {
	pubKeyBytes, err := pubKeyHex.Bytes()
	if err != nil {
		return nil, fmt.Errorf("parsing from hex: %w", err)
	}
	pubKey, err := x509.ParsePKCS1PublicKey(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("parsing from PKCS1: %w", err)
	}
	return pubKey, nil
}

// EncryptSeedShares encrypts a seed for owners identified by their public keys and returns a SeedShare slice suitable for userapi.SetManifestResponse.
func EncryptSeedShares(seed []byte, ownerPubKeys []HexString) ([]*userapi.SeedShare, error) {
	var out []*userapi.SeedShare
	for _, pubKeyHex := range ownerPubKeys {
		pubKey, err := ParseSeedShareOwnerKey(pubKeyHex)
		if err != nil {
			return nil, fmt.Errorf("parsing seed share owner key: %w", err)
		}
		cipherText, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, seed, []byte("seedshare"))
		if err != nil {
			return nil, fmt.Errorf("encrypting seed share: %w", err)
		}
		seedShare := &userapi.SeedShare{
			EncryptedSeed: cipherText,
			PublicKey:     pubKeyHex.String(),
		}
		out = append(out, seedShare)
	}
	return out, nil
}

// DecryptSeedShare tries to decrypt a SeedShare with the given owner key.
func DecryptSeedShare(key *rsa.PrivateKey, seedShare *userapi.SeedShare) ([]byte, error) {
	// TODO(burgerdev): check seedShare.PublicKey?
	return rsa.DecryptOAEP(sha256.New(), nil, key, seedShare.GetEncryptedSeed(), []byte("seedshare"))
}

// NewWorkloadOwnerKey creates and marshals a private key.
func NewWorkloadOwnerKey() ([]byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating private key: %w", err)
	}
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("marshaling private key: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKeyBytes}), nil
}
