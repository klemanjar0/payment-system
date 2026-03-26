package http

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/klemanjar0/payment-system/pkg/httputil"
	"github.com/klemanjar0/payment-system/services/transaction/internal/domain"
	usecase "github.com/klemanjar0/payment-system/services/transaction/internal/use_case"
)

type TransactionHTTPHandler struct {
	creator usecase.TransactionCreator
	getter  usecase.TransactionGetter
	lister  usecase.TransactionsByAccountGetter
}

func NewTransactionHTTPHandler(
	creator usecase.TransactionCreator,
	getter usecase.TransactionGetter,
	lister usecase.TransactionsByAccountGetter,
) *TransactionHTTPHandler {
	return &TransactionHTTPHandler{
		creator: creator,
		getter:  getter,
		lister:  lister,
	}
}

// --- request / response structs ---

type createTransferRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
	FromAccountID  string `json:"from_account_id"`
	ToAccountID    string `json:"to_account_id"`
	Amount         int64  `json:"amount"`
	Currency       string `json:"currency"`
	Description    string `json:"description"`
}

type createTransferResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}

type transactionResponse struct {
	ID             string  `json:"id"`
	IdempotencyKey string  `json:"idempotency_key"`
	FromAccountID  string  `json:"from_account_id"`
	ToAccountID    string  `json:"to_account_id"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	Description    string  `json:"description"`
	Status         string  `json:"status"`
	FailureReason  string  `json:"failure_reason,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type transactionListResponse struct {
	Transactions []transactionResponse `json:"transactions"`
	Total        int                   `json:"total"`
}

// --- handlers ---

func (h *TransactionHTTPHandler) CreateTransfer(c fiber.Ctx) error {
	var req createTransferRequest
	if err := c.Bind().Body(&req); err != nil {
		return httputil.Respond(c).BadRequest(err)
	}

	tx, err := h.creator.Execute(c.Context(), &usecase.CreateTransferParams{
		IdempotencyKey: req.IdempotencyKey,
		FromAccountID:  req.FromAccountID,
		ToAccountID:    req.ToAccountID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Description:    req.Description,
	})
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}

	return httputil.Respond(c).Created(createTransferResponse{
		TransactionID: tx.ID,
		Status:        string(tx.Status),
	})
}

func (h *TransactionHTTPHandler) GetTransaction(c fiber.Ctx) error {
	tx, err := h.getter.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}
	return httputil.Respond(c).OK(domainTransactionToResponse(tx))
}

func (h *TransactionHTTPHandler) GetTransactionsByAccount(c fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	txs, total, err := h.lister.Execute(c.Context(), c.Params("account_id"), limit, offset)
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}

	items := make([]transactionResponse, len(txs))
	for i, tx := range txs {
		items[i] = domainTransactionToResponse(tx)
	}
	return httputil.Respond(c).OK(transactionListResponse{
		Transactions: items,
		Total:        total,
	})
}

// --- mapper ---

func domainTransactionToResponse(tx *domain.Transaction) transactionResponse {
	return transactionResponse{
		ID:             tx.ID,
		IdempotencyKey: tx.IdempotencyKey,
		FromAccountID:  tx.FromAccountID,
		ToAccountID:    tx.ToAccountID,
		Amount:         tx.Amount.ToMajor(),
		Currency:       tx.Currency,
		Description:    tx.Description,
		Status:         string(tx.Status),
		FailureReason:  tx.FailureReason,
		CreatedAt:      tx.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      tx.UpdatedAt.Format(time.RFC3339),
	}
}
