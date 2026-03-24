package domain

import (
	"time"

	"github.com/klemanjar0/payment-system/pkg/logger"
	"github.com/klemanjar0/payment-system/services/user/internal/repository/postgres/sqlc"
)

type RefreshToken struct {
	TokenID     string
	UserID      string
	DeviceInfo  string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	LastUsedAt  time.Time
	RotatedFrom string
	Revoked     bool
}

func NewRefreshTokenOfSql(tkn *sqlc.RefreshToken) (*RefreshToken, error) {
	id := tkn.ID.String()

	if id == "" {
		logger.Error("user ID is missing")
		return nil, ErrInternal
	}

	return &RefreshToken{
		TokenID:     id,
		UserID:      tkn.UserID.String(),
		DeviceInfo:  tkn.DeviceInfo.String,
		CreatedAt:   tkn.CreatedAt.Time,
		ExpiresAt:   tkn.ExpiresAt.Time,
		RotatedFrom: tkn.RotatedFrom.String(),
		Revoked:     tkn.Revoked.Bool,
		LastUsedAt:  tkn.LastUsedAt.Time,
	}, nil
}

func (t *RefreshToken) IsRevoked() bool {
	return t.Revoked
}
