package jwtx

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrMissingBearer = errors.New("missing token")
	ErrInvalidToken  = errors.New("mnvalid token")
)

type TokenManager struct {
	Secret []byte
	Issuer string
}

type AccessClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func New(secret, issuer string) *TokenManager {
	return &TokenManager{
		Secret: []byte(secret),
		Issuer: issuer,
	}
}

func (m *TokenManager) SignToken(userID, email, role string) (string, error) {
	now := time.Now().UTC()
	claims := AccessClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   m.Issuer,
			Subject:  userID,
			IssuedAt: jwt.NewNumericDate(now),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return t.SignedString(m.Secret)
}

func (m *TokenManager) VerifyToken(token string) (*AccessClaims, error) {
	claims := new(AccessClaims)
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}

		return m.Secret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func ExtractToken(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")

	if h == "" {
		return "", ErrMissingBearer
	}

	parts := strings.SplitN(h, " ", 2)

	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", ErrMissingBearer
	}

	return parts[1], nil
}
