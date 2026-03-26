package http

import "github.com/gofiber/fiber/v3"

// RegisterRoutes wires all account HTTP routes onto the given Fiber app.
// All routes require a valid Bearer access token validated by authMiddleware.
func RegisterRoutes(app *fiber.App, h *AccountHTTPHandler, authMiddleware fiber.Handler) {
	v1 := app.Group("/v1", authMiddleware)

	v1.Post("/accounts", h.CreateAccount)
	v1.Get("/accounts/me", h.GetMyAccounts)
	v1.Get("/accounts/user/:user_id", h.GetAccountsByUser)
	v1.Get("/accounts/:id", h.GetAccount)
}
