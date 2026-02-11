package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/klemanjar0/payment-system/pkg/logger"
	utilid "github.com/klemanjar0/payment-system/pkg/util_id"
	"github.com/klemanjar0/payment-system/services/user/internal/domain"
	"github.com/klemanjar0/payment-system/services/user/internal/repository/postgres/sqlc"
)

type UserRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	queries := sqlc.New(pool)
	return &UserRepository{pool: pool, queries: queries}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	entity, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        user.Email,
		Phone:        user.Phone,
		PasswordHash: user.PasswordHash,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Status:       sqlc.UserStatus(user.Status),
		KycStatus:    sqlc.KycStatus(user.KYCStatus),
		CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})

	if err == nil {
		logger.Debug("UserRepository->Create. Entity Created", "User", entity)
	} else {
		return nil, err
	}

	return domain.NewUserOfSql(&entity)
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	entity, err := r.queries.GetByID(ctx, utilid.FromString(id).AsPgUUID())

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return domain.NewUserOfSql(&entity)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	entity, err := r.queries.GetByEmail(ctx, email)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return domain.NewUserOfSql(&entity)
}

func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	entity, err := r.queries.GetByPhone(ctx, phone)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return domain.NewUserOfSql(&entity)
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	entity, err := r.queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:           utilid.FromString(user.ID).AsPgUUID(),
		Email:        user.Email,
		Phone:        user.Phone,
		PasswordHash: user.PasswordHash,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Status:       sqlc.UserStatus(user.Status),
		KycStatus:    sqlc.KycStatus(user.KYCStatus),
		UpdatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return domain.NewUserOfSql(&entity)
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	exists, err := r.queries.ExistsByEmail(ctx, email)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, domain.ErrUserNotFound
	}

	if err != nil {
		return false, err
	}

	return exists, nil
}
