package accounts

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	AccountID string `json:"uid"`
	Email     string `json:"email"`
	TenantID  string `json:"tid"`
	jwt.RegisteredClaims
}

type TokenService struct {
	secret []byte
	ttl    time.Duration
	tenant string
}

func NewTokenService(secret string, ttl time.Duration, tenantID string) *TokenService {
	return &TokenService{
		secret: []byte(secret),
		ttl:    ttl,
		tenant: tenantID,
	}
}

func (s *TokenService) Issue(account Account, tenantID string) (Session, error) {
	tid := tenantID
	if tid == "" {
		tid = s.tenant
	}
	now := time.Now()
	claims := Claims{
		AccountID: account.ID,
		Email:     account.Email,
		TenantID:  tid,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   account.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return Session{}, fmt.Errorf("sign token: %w", err)
	}

	return Session{
		AccessToken: signed,
		Account:     account,
		TenantID:    tid,
	}, nil
}

func (s *TokenService) Parse(accessToken string) (Claims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return Claims{}, ErrInvalidToken
	}
	return *claims, nil
}
