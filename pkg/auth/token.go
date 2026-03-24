package auth

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrSigningNotAvailable = errors.New("auth: signing not available in validator-only mode")
	ErrPublicKeyRequired   = errors.New("auth: public key is required")
	ErrPrivateKeyRequired  = errors.New("auth: private key is required for token service")
	ErrInvalidToken        = errors.New("auth: invalid or expired token")
	ErrTokenTypeMismatch   = errors.New("auth: token type mismatch")
	ErrExpiredToken        = errors.New("token has expired")
)

type tokenService struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	config     Config
	canSign    bool
}

// NewTokenService creates a TokenService capable of both signing and verifying tokens.
// Requires both private and public keys in the Config.
func NewTokenService(cfg Config) (TokenService, error) {
	if cfg.PrivateKey == nil {
		return nil, ErrPrivateKeyRequired
	}
	if cfg.PublicKey == nil {
		cfg.PublicKey = DerivePublicKey(cfg.PrivateKey)
	}
	cfg.applyDefaults()

	return &tokenService{
		privateKey: cfg.PrivateKey,
		publicKey:  cfg.PublicKey,
		config:     cfg,
		canSign:    true,
	}, nil
}

// NewTokenValidator creates a TokenService that can only verify tokens.
// Generate methods will return ErrSigningNotAvailable.
func NewTokenValidator(cfg Config) (TokenService, error) {
	if cfg.PublicKey == nil {
		return nil, ErrPublicKeyRequired
	}
	cfg.applyDefaults()

	return &tokenService{
		publicKey: cfg.PublicKey,
		config:    cfg,
		canSign:   false,
	}, nil
}

func (s *tokenService) GenerateAccessToken(userID string) (string, error) {
	return s.generateToken(userID, AccessToken, s.config.AccessTokenExpiry)
}

func (s *tokenService) GenerateRefreshToken(userID string) (string, error) {
	return s.generateToken(userID, RefreshToken, s.config.RefreshTokenExpiry)
}

func (s *tokenService) ValidateAccessToken(token string) (*Claims, error) {
	return s.validateToken(token, AccessToken)
}

func (s *tokenService) ValidateRefreshToken(token string) (*Claims, error) {
	return s.validateToken(token, RefreshToken)
}

func (s *tokenService) generateToken(tokenId string, tokenType TokenType, expiry time.Duration) (string, error) {
	if !s.canSign {
		return "", ErrSigningNotAvailable
	}

	now := time.Now()
	claims := Claims{
		TokenID:   tokenId,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    s.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(s.privateKey)
}

func (s *tokenService) validateToken(tokenString string, expectedType TokenType) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("auth: unexpected signing method: %v", t.Header["alg"])
		}
		return s.publicKey, nil
	},
		jwt.WithIssuer(s.config.Issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != expectedType {
		return nil, ErrTokenTypeMismatch
	}

	return claims, nil
}
