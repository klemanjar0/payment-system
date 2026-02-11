package domain

import (
	"time"
	"unicode"

	"github.com/klemanjar0/payment-system/pkg/logger"
	"github.com/klemanjar0/payment-system/services/user/internal/repository/postgres/sqlc"
	"golang.org/x/crypto/bcrypt"
)

type UserStatus string

const (
	UserStatusPending UserStatus = "pending"
	UserStatusActive  UserStatus = "active"
	UserStatusBlocked UserStatus = "blocked"
	UserStatusDeleted UserStatus = "deleted"
)

type KYCStatus string

const (
	KYCStatusNone     KYCStatus = "none"
	KYCStatusPending  KYCStatus = "pending"
	KYCStatusVerified KYCStatus = "verified"
	KYCStatusRejected KYCStatus = "rejected"
)

type User struct {
	ID           string
	Email        string
	Phone        string
	PasswordHash string
	FirstName    string
	LastName     string
	Status       UserStatus
	KYCStatus    KYCStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(email, phone, password, firstName, lastName string) (*User, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	if err := validatePassword(password); err != nil {
		return nil, err
	}

	if err := validatePhone(phone); err != nil {
		return nil, err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return nil, ErrInternal
	}

	now := time.Now()

	return &User{
		Email:        email,
		Phone:        normalizePhone(phone),
		PasswordHash: hash,
		FirstName:    firstName,
		LastName:     lastName,
		Status:       UserStatusPending,
		KYCStatus:    KYCStatusNone,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func NewUserOfSql(usr *sqlc.User) (*User, error) {
	id := usr.ID.String()

	if id == "" {
		logger.Error("user ID is missing")
		return nil, ErrInternal
	}

	return &User{
		ID:           id,
		Email:        usr.Email,
		Phone:        usr.Phone,
		PasswordHash: usr.PasswordHash,
		FirstName:    usr.FirstName,
		LastName:     usr.LastName,
		Status:       UserStatus(usr.Status),
		KYCStatus:    KYCStatus(usr.KycStatus),
		CreatedAt:    usr.CreatedAt.Time,
		UpdatedAt:    usr.UpdatedAt.Time,
	}, nil
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

func (u *User) CanTransact() bool {
	return u.IsActive() && u.KYCStatus == KYCStatusVerified
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) ChangePassword(oldPassword, newPassword string) error {
	if !u.CheckPassword(oldPassword) {
		return ErrInvalidCredentials
	}

	if err := validatePassword(newPassword); err != nil {
		return err
	}

	hash, err := hashPassword(newPassword)
	if err != nil {
		return ErrInternal
	}

	u.PasswordHash = hash
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) UpdateProfile(firstName, lastName, phone *string) error {
	if firstName != nil {
		u.FirstName = *firstName
	}
	if lastName != nil {
		u.LastName = *lastName
	}
	if phone != nil {
		if err := validatePhone(*phone); err != nil {
			return err
		}
		u.Phone = normalizePhone(*phone)
	}

	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) Activate() {
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now()
}

func (u *User) Block() {
	u.Status = UserStatusBlocked
	u.UpdatedAt = time.Now()
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func validateEmail(email string) error {
	if len(email) < 5 || len(email) > 255 {
		return ErrInvalidEmail
	}

	hasAt := false
	hasDot := false
	for i, c := range email {
		if c == '@' {
			if hasAt || i == 0 {
				return ErrInvalidEmail
			}
			hasAt = true
		}
		if c == '.' && hasAt {
			hasDot = true
		}
	}
	if !hasAt || !hasDot {
		return ErrInvalidEmail
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	if len(password) > 128 {
		return ErrPasswordTooLong
	}

	var hasUpper, hasLower, hasDigit bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return ErrPasswordTooWeak
	}

	return nil
}

func validatePhone(phone string) error {
	digits := 0
	for _, c := range phone {
		if unicode.IsDigit(c) {
			digits++
		}
	}
	if digits < 10 || digits > 15 {
		return ErrInvalidPhone
	}
	return nil
}

func normalizePhone(phone string) string {
	var result []rune
	for _, c := range phone {
		if unicode.IsDigit(c) {
			result = append(result, c)
		}
	}
	return string(result)
}
