package http

import "github.com/gofiber/fiber/v3"

// RegisterRoutes wires all transaction HTTP routes onto the given Fiber app.
// All routes require a valid Bearer access token validated by authMiddleware.
func RegisterRoutes(app *fiber.App, h *TransactionHTTPHandler, authMiddleware fiber.Handler) {
	v1 := app.Group("/v1", authMiddleware)

	v1.Post("/transfers", h.CreateTransfer)
	v1.Get("/transactions/:id", h.GetTransaction)
	v1.Get("/accounts/:account_id/transactions", h.GetTransactionsByAccount)
}
