package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/klemanjar0/payment-system/pkg/auth"
	"github.com/klemanjar0/payment-system/pkg/httputil"
	"github.com/klemanjar0/payment-system/pkg/tokenblacklist"
)

// Auth returns a Fiber middleware that validates a Bearer JWT access token.
// On success it stores userID and the raw token string in c.Locals.
func Auth(tokenSvc auth.TokenService, bl tokenblacklist.Blacklist) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return httputil.Respond(c).Unauthorized(auth.ErrInvalidToken)
		}

		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := tokenSvc.ValidateAccessToken(token)
		if err != nil {
			if errors.Is(err, auth.ErrExpiredToken) {
				return httputil.Respond(c).Unauthorized(auth.ErrExpiredToken)
			}
			return httputil.Respond(c).Unauthorized(auth.ErrInvalidToken)
		}

		// Reject tokens that have been blacklisted (e.g. after logout).
		if blacklisted, _ := bl.IsBlacklisted(c.Context(), claims.ID); blacklisted {
			return httputil.Respond(c).Unauthorized(auth.ErrInvalidToken)
		}

		c.Locals("userID", claims.TokenID)
		c.Locals("accessToken", token) // raw token string, used by logout handler
		return c.Next()
	}
}
