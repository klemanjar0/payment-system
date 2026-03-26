package http

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/klemanjar0/payment-system/pkg/httputil"
	"github.com/klemanjar0/payment-system/services/account/internal/domain"
	usecase "github.com/klemanjar0/payment-system/services/account/internal/use_case"
)

type AccountHTTPHandler struct {
	creator      usecase.AccountCreator
	getter       usecase.AccountGetter
	listerByUser usecase.AccountsByUserGetter
}

func NewAccountHTTPHandler(
	creator usecase.AccountCreator,
	getter usecase.AccountGetter,
	listerByUser usecase.AccountsByUserGetter,
) *AccountHTTPHandler {
	return &AccountHTTPHandler{
		creator:      creator,
		getter:       getter,
		listerByUser: listerByUser,
	}
}

// --- request / response structs ---

type createAccountRequest struct {
	Currency string `json:"currency"`
}

type accountResponse struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	Currency    string  `json:"currency"`
	Balance     float64 `json:"balance"`
	HoldAmount  float64 `json:"hold_amount"`
	Available   float64 `json:"available"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type accountListResponse struct {
	Accounts []accountResponse `json:"accounts"`
}

// --- handlers ---

func (h *AccountHTTPHandler) CreateAccount(c fiber.Ctx) error {
	var req createAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return httputil.Respond(c).BadRequest(err)
	}

	userID, _ := c.Locals("userID").(string)

	account, err := h.creator.Execute(c.Context(), &usecase.CreateAccountUseCaseParams{
		UserID:   userID,
		Currency: domain.Currency(req.Currency),
	})
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}

	return httputil.Respond(c).Created(domainAccountToResponse(account))
}

func (h *AccountHTTPHandler) GetAccount(c fiber.Ctx) error {
	account, err := h.getter.Execute(c.Context(), c.Params("id"))
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}
	return httputil.Respond(c).OK(domainAccountToResponse(account))
}

func (h *AccountHTTPHandler) GetMyAccounts(c fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	return h.listByUserID(c, userID)
}

func (h *AccountHTTPHandler) GetAccountsByUser(c fiber.Ctx) error {
	return h.listByUserID(c, c.Params("user_id"))
}

func (h *AccountHTTPHandler) listByUserID(c fiber.Ctx, userID string) error {
	accounts, err := h.listerByUser.Execute(c.Context(), userID)
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}

	list := make([]accountResponse, len(accounts))
	for i, a := range accounts {
		list[i] = domainAccountToResponse(a)
	}
	return httputil.Respond(c).OK(accountListResponse{Accounts: list})
}

// --- mapper ---

func domainAccountToResponse(a *domain.Account) accountResponse {
	return accountResponse{
		ID:         a.ID,
		UserID:     a.UserID,
		Currency:   string(a.Currency),
		Balance:    a.Balance.ToMajor(),
		HoldAmount: a.HoldAmount.ToMajor(),
		Available:  a.Available().ToMajor(),
		Status:     string(a.Status),
		CreatedAt:  a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  a.UpdatedAt.Format(time.RFC3339),
	}
}
