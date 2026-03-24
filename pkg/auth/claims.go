package auth

import "github.com/golang-jwt/jwt/v5"

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	TokenID   string    `json:"token_id"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenService interface {
	GenerateAccessToken(userID string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateAccessToken(token string) (*Claims, error)
	ValidateRefreshToken(token string) (*Claims, error)
}
