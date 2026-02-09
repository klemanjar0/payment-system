package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/klemanjar0/payment-system/services/user/internal/domain"
)

type CreateUserUseCase struct {
	repo     domain.UserRepository
	tokenSvc TokenService
	eventPub EventPublisher
}

func NewCreateUserUseCase(
	repo domain.UserRepository,
	tokenSvc TokenService,
	eventPub EventPublisher,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		repo:     repo,
		tokenSvc: tokenSvc,
		eventPub: eventPub,
	}
}

type CreateUserInput struct {
	Email     string
	Phone     string
	Password  string
	FirstName string
	LastName  string
}

type CreateUserOutput struct {
	UserID       string
	AccessToken  string
	RefreshToken string
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
	exists, err := uc.repo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, domain.ErrInternal
	}
	if exists {
		return nil, domain.ErrUserAlreadyExists
	}

	user, err := domain.NewUser(
		input.Email,
		input.Phone,
		input.Password,
		input.FirstName,
		input.LastName,
	)
	if err != nil {
		return nil, err
	}

	user.ID = uuid.New().String()

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, domain.ErrInternal
	}

	accessToken, err := uc.tokenSvc.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	refreshToken, err := uc.tokenSvc.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	uc.eventPub.Publish(ctx, "user.created", UserCreatedEvent{
		UserID:    user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})

	return &CreateUserOutput{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
