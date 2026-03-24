package auth

import (
	"crypto/ecdsa"
	"time"
)

const (
	DefaultAccessTokenExpiry  = 15 * time.Minute
	DefaultRefreshTokenExpiry = 7 * 24 * time.Hour
	DefaultIssuer             = "payment-system"
)

type Config struct {
	// PrivateKey is the ECDSA private key used for signing tokens.
	// Required for NewTokenService; nil for NewTokenValidator (verify-only mode).
	PrivateKey *ecdsa.PrivateKey

	// PublicKey is the ECDSA public key used for verifying token signatures.
	// Required for both NewTokenService and NewTokenValidator.
	PublicKey *ecdsa.PublicKey

	// AccessTokenExpiry is the lifetime of access tokens. Default: 15 minutes.
	AccessTokenExpiry time.Duration

	// RefreshTokenExpiry is the lifetime of refresh tokens. Default: 7 days.
	RefreshTokenExpiry time.Duration

	// Issuer is the "iss" claim value. Default: "payment-system".
	Issuer string
}

func (c *Config) applyDefaults() {
	if c.AccessTokenExpiry == 0 {
		c.AccessTokenExpiry = DefaultAccessTokenExpiry
	}
	if c.RefreshTokenExpiry == 0 {
		c.RefreshTokenExpiry = DefaultRefreshTokenExpiry
	}
	if c.Issuer == "" {
		c.Issuer = DefaultIssuer
	}
}
