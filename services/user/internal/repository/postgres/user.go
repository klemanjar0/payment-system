package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/klemanjar0/payment-system/services/user/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, phone, password_hash, first_name, last_name, status, kyc_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Status,
		user.KYCStatus,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, phone, password_hash, first_name, last_name, status, kyc_status, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	return r.scanUser(ctx, query, id)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, phone, password_hash, first_name, last_name, status, kyc_status, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	return r.scanUser(ctx, query, email)
}

func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
		SELECT id, email, phone, password_hash, first_name, last_name, status, kyc_status, created_at, updated_at
		FROM users
		WHERE phone = $1
	`

	return r.scanUser(ctx, query, phone)
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $2, phone = $3, password_hash = $4, first_name = $5, last_name = $6, 
		    status = $7, kyc_status = $8, updated_at = $9
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Status,
		user.KYCStatus,
		user.UpdatedAt,
	)

	return err
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *UserRepository) scanUser(ctx context.Context, query string, arg interface{}) (*domain.User, error) {
	var user domain.User

	err := r.pool.QueryRow(ctx, query, arg).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Status,
		&user.KYCStatus,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
