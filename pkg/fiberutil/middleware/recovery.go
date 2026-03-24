package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/klemanjar0/payment-system/pkg/logger"
)

// Recovery returns a Fiber middleware that catches panics and returns 500.
func Recovery() fiber.Handler {
	return func(c fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					"method", c.Method(),
					"path", c.Path(),
					"panic", r,
				)
				err = c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"error": "internal server error",
					"code":  http.StatusInternalServerError,
				})
			}
		}()
		return c.Next()
	}
}
