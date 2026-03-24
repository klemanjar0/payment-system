package http

import "github.com/gofiber/fiber/v3"

// RegisterRoutes wires all user HTTP routes onto the given Fiber app.
// Protected routes require a valid Bearer access token validated by authMiddleware.
func RegisterRoutes(app *fiber.App, h *UserHTTPHandler, authMiddleware fiber.Handler) {
	v1 := app.Group("/v1")

	// public
	v1.Post("/users", h.Register)
	v1.Post("/auth/login", h.Authenticate)

	// protected
	users := v1.Group("/users", authMiddleware)
	users.Get("/me", h.Me)
	users.Get("/:id", h.GetUser)
	users.Get("/email/:email", h.GetUserByEmail)
	users.Get("/:id/validate", h.ValidateUser)
	users.Post("/:id/change-password", h.ChangePassword)
}
