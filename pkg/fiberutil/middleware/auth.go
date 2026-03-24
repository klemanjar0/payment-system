package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/klemanjar0/payment-system/pkg/auth"
	"github.com/klemanjar0/payment-system/pkg/httputil"
)

// Auth returns a Fiber middleware that validates a Bearer JWT access token.
// On success, it stores the userID in c.Locals("userID").
func Auth(tokenSvc auth.TokenService) fiber.Handler {
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

		c.Locals("userID", claims.TokenID)
		return c.Next()
	}
}
