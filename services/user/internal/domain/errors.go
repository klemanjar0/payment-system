package domain

import "errors"

var (
	// user errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserBlocked        = errors.New("user is blocked")
	ErrUserNotActive      = errors.New("user is not active")

	// validation errors
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrInvalidPhone     = errors.New("invalid phone format")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong  = errors.New("password must be less than 128 characters")
	ErrPasswordTooWeak  = errors.New("password must contain uppercase, lowercase and digit")

	// generic
	ErrInternal = errors.New("internal error")
)
