package http

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/klemanjar0/payment-system/pkg/httputil"
	"github.com/klemanjar0/payment-system/services/user/internal/domain"
	"github.com/klemanjar0/payment-system/services/user/internal/usecase"
)

type UserHTTPHandler struct {
	createUser     *usecase.CreateUserUseCase
	authenticate   *usecase.AuthenticateUseCase
	getUser        *usecase.GetUserUseCase
	changePassword *usecase.ChangePasswordUseCase
}

func NewUserHTTPHandler(
	createUser *usecase.CreateUserUseCase,
	authenticate *usecase.AuthenticateUseCase,
	getUser *usecase.GetUserUseCase,
	changePassword *usecase.ChangePasswordUseCase,
) *UserHTTPHandler {
	return &UserHTTPHandler{
		createUser:     createUser,
		authenticate:   authenticate,
		getUser:        getUser,
		changePassword: changePassword,
	}
}

// --- request / response structs ---

type registerRequest struct {
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type registerResponse struct {
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	CreatedAt    time.Time `json:"created_at"`
}

type authenticateRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authenticateResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type userResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Status    string `json:"status"`
	KYCStatus string `json:"kyc_status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type validateUserResponse struct {
	Valid     bool   `json:"valid"`
	Status    string `json:"status"`
	KYCStatus string `json:"kyc_status"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// --- handlers ---

func (h *UserHTTPHandler) Register(c fiber.Ctx) error {
	var req registerRequest
	if err := c.Bind().Body(&req); err != nil {
		return httputil.Respond(c).BadRequest(err)
	}

	result, err := h.createUser.Execute(c.Context(), usecase.CreateUserInput{
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}

	return httputil.Respond(c).Created(registerResponse{
		UserID:       result.UserID,
		Email:        result.Email,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		CreatedAt:    result.CreatedAt,
	})
}

func (h *UserHTTPHandler) Authenticate(c fiber.Ctx) error {
	var req authenticateRequest
	if err := c.Bind().Body(&req); err != nil {
		return httputil.Respond(c).BadRequest(err)
	}

	result, err := h.authenticate.Execute(c.Context(), usecase.AuthenticateInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}

	return httputil.Respond(c).OK(authenticateResponse{
		UserID:       result.UserID,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}

func (h *UserHTTPHandler) GetUser(c fiber.Ctx) error {
	user, err := h.getUser.ExecuteByID(c.Context(), c.Params("id"))
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}
	return httputil.Respond(c).OK(domainUserToResponse(user))
}

func (h *UserHTTPHandler) GetUserByEmail(c fiber.Ctx) error {
	user, err := h.getUser.ExecuteByEmail(c.Context(), c.Params("email"))
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}
	return httputil.Respond(c).OK(domainUserToResponse(user))
}

func (h *UserHTTPHandler) ValidateUser(c fiber.Ctx) error {
	user, err := h.getUser.ExecuteByID(c.Context(), c.Params("id"))
	if err != nil {
		return httputil.Respond(c).OK(validateUserResponse{Valid: false})
	}
	return httputil.Respond(c).OK(validateUserResponse{
		Valid:     user.IsActive(),
		Status:    string(user.Status),
		KYCStatus: string(user.KYCStatus),
	})
}

func (h *UserHTTPHandler) Me(c fiber.Ctx) error {
	userID, _ := c.Locals("userID").(string)
	user, err := h.getUser.ExecuteByID(c.Context(), userID)
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}
	return httputil.Respond(c).OK(domainUserToResponse(user))
}

func (h *UserHTTPHandler) ChangePassword(c fiber.Ctx) error {
	var req changePasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return httputil.Respond(c).BadRequest(err)
	}

	err := h.changePassword.Execute(c.Context(), usecase.ChangePasswordInput{
		UserID:      c.Params("id"),
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		return httputil.Respond(c).Error(err).Send()
	}

	return httputil.Respond(c).NoContent()
}

// --- mapper ---

func domainUserToResponse(u *domain.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Email:     u.Email,
		Phone:     u.Phone,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Status:    string(u.Status),
		KYCStatus: string(u.KYCStatus),
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
}
