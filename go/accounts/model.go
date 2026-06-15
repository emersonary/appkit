package accounts

import "time"

type Account struct {
	ID              string
	Email           string
	PasswordHash    *string
	FirstName       *string
	LastName        *string
	AvatarURL       *string
	EmailVerifiedAt *time.Time
	IsAdmin         bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (a Account) EmailVerified() bool {
	return a.EmailVerifiedAt != nil
}

type Session struct {
	AccessToken string
	Account     Account
	TenantID    string
}

type RegisterResult struct {
	VerificationRequired bool
	Account              Account
}

type VerifyEmailResult struct {
	Email           string
	AlreadyVerified bool
}

type Tenant struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
