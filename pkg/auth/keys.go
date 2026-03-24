package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

var (
	ErrInvalidPEMBlock  = errors.New("auth: failed to decode PEM block")
	ErrInvalidKeyFormat = errors.New("auth: unsupported key format")
)

// LoadPrivateKeyFromFile reads a PEM-encoded ECDSA private key from the given file path.
func LoadPrivateKeyFromFile(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadPrivateKeyFromPEM(data)
}

// LoadPublicKeyFromFile reads a PEM-encoded ECDSA public key from the given file path.
func LoadPublicKeyFromFile(path string) (*ecdsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadPublicKeyFromPEM(data)
}

// LoadPrivateKeyFromPEM parses a PEM-encoded ECDSA private key.
func LoadPrivateKeyFromPEM(data []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, ErrInvalidPEMBlock
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format as fallback.
		parsed, pkcs8Err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if pkcs8Err != nil {
			return nil, err
		}
		ecKey, ok := parsed.(*ecdsa.PrivateKey)
		if !ok {
			return nil, ErrInvalidKeyFormat
		}
		return ecKey, nil
	}
	return key, nil
}

// LoadPublicKeyFromPEM parses a PEM-encoded ECDSA public key.
func LoadPublicKeyFromPEM(data []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, ErrInvalidPEMBlock
	}

	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	ecKey, ok := parsed.(*ecdsa.PublicKey)
	if !ok {
		return nil, ErrInvalidKeyFormat
	}
	return ecKey, nil
}

// DerivePublicKey extracts the public key from an ECDSA private key.
func DerivePublicKey(priv *ecdsa.PrivateKey) *ecdsa.PublicKey {
	return &priv.PublicKey
}
