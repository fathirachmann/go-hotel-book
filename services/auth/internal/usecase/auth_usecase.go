package usecase

import (
	"context"
	"errors"
	"strings"

	"auth/internal/domain"
	"auth/internal/repo"

	"pkg/bcryptx"
	"pkg/jwtx"

	"gorm.io/gorm"
)

const defaultRole = "customer"

var (
	ErrEmailAlreadyUsed   = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type RegisterInput struct {
	FullName string
	Email    string
	Password string
	Role     string
}

type LoginInput struct {
	Email    string
	Password string
}

// AuthResult packs user-facing fields and freshly issued access token.
type AuthResult struct {
	AccessToken string       `json:"access_token"`
	User        *UserPayload `json:"user"`
}

// UserPayload is a thin representation of the account exposed through APIs.
type UserPayload struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// AuthUsecase defines the behaviour needed by HTTP handlers.
type AuthUsecase interface {
	Register(ctx context.Context, input RegisterInput) (*AuthResult, error)
	Login(ctx context.Context, input LoginInput) (*AuthResult, error)
}

type authUsecase struct {
	repo   repo.UserRepository
	tokens *jwtx.TokenManager
}

// NewAuthUsecase wires repository plus token manager into an AuthUsecase.
func NewAuthUsecase(repo repo.UserRepository, tokens *jwtx.TokenManager) AuthUsecase {
	return &authUsecase{repo: repo, tokens: tokens}
}

func (uc *authUsecase) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" || strings.TrimSpace(input.Password) == "" || strings.TrimSpace(input.FullName) == "" {
		return nil, ErrInvalidCredentials
	}

	if _, err := uc.repo.FindByEmail(ctx, email); err == nil {
		return nil, ErrEmailAlreadyUsed
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashed, err := bcryptx.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	role := strings.TrimSpace(input.Role)
	if role == "" {
		role = defaultRole
	}

	user := &domain.User{
		FullName:       strings.TrimSpace(input.FullName),
		Email:          email,
		HashedPassword: hashed,
		Role:           role,
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := uc.tokens.SignToken(user.ID.String(), user.Role)
	if err != nil {
		return nil, err
	}

	return buildAuthResult(user, token), nil
}

func (uc *authUsecase) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" || strings.TrimSpace(input.Password) == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcryptx.CompareHash(user.HashedPassword, input.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := uc.tokens.SignToken(user.ID.String(), user.Role)
	if err != nil {
		return nil, err
	}

	return buildAuthResult(user, token), nil
}

func buildAuthResult(user *domain.User, token string) *AuthResult {
	return &AuthResult{
		AccessToken: token,
		User: &UserPayload{
			ID:       user.ID.String(),
			FullName: user.FullName,
			Email:    user.Email,
			Role:     user.Role,
		},
	}
}
