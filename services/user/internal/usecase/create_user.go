package usecase

import (
	"context"
	"time"

	"github.com/klemanjar0/payment-system/pkg/auth"
	"github.com/klemanjar0/payment-system/services/user/internal/domain"
)

type CreateUserInput struct {
	Email     string
	Phone     string
	Password  string
	FirstName string
	LastName  string
}

type CreateUserUseCase struct {
	repo      domain.UserRepository
	tokenSvc  auth.TokenService
	eventPub  EventPublisher
	auditLog  UserAuditLogger
}

type CreateUserResult struct {
	UserID       string
	Email        string
	AccessToken  string
	RefreshToken string
	CreatedAt    time.Time
}

func NewCreateUserUseCase(
	repo domain.UserRepository,
	tokenSvc auth.TokenService,
	eventPub EventPublisher,
	auditLog UserAuditLogger,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		repo:     repo,
		tokenSvc: tokenSvc,
		eventPub: eventPub,
		auditLog: auditLog,
	}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*CreateUserResult, error) {
	exists, err := uc.repo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, domain.ErrInternal
	}
	if exists {
		return nil, domain.ErrUserAlreadyExists
	}

	data, err := domain.NewUser(
		input.Email,
		input.Phone,
		input.Password,
		input.FirstName,
		input.LastName,
	)
	if err != nil {
		return nil, err
	}

	newUser, err := uc.repo.Create(ctx, data)
	if err != nil {
		return nil, domain.ErrInternal
	}

	accessToken, err := uc.tokenSvc.GenerateAccessToken(newUser.ID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	refreshToken, err := uc.tokenSvc.GenerateRefreshToken(newUser.ID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	uc.eventPub.Publish(ctx, "user.created", UserCreatedEvent{
		UserID:    newUser.ID,
		Email:     newUser.Email,
		CreatedAt: newUser.CreatedAt,
	})

	uc.auditLog.LogUserCreated(ctx, newUser.ID, newUser.Email)

	return &CreateUserResult{
		UserID:       newUser.ID,
		Email:        newUser.Email,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		CreatedAt:    newUser.CreatedAt,
	}, nil
}
