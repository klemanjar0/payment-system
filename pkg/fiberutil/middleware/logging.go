package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/klemanjar0/payment-system/pkg/logger"
)

// Logging returns a Fiber middleware that logs each HTTP request.
func Logging() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		logger.Info("http request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return err
	}
}
