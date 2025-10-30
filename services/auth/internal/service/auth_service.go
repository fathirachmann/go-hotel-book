package service

import (
	"context"
	"errors"
	"strings"

	"auth/internal/entity"
	"auth/internal/repo"

	"pkg/bcryptx"
	"pkg/jwtx"

	"gorm.io/gorm"
)

const (
	RoleUser  = "USER"
	RoleStaff = "STAFF"
	RoleAdmin = "ADMIN"
)

var allowedRoles = map[string]struct{}{
	RoleUser:  {},
	RoleStaff: {},
	RoleAdmin: {},
}

const defaultRole = RoleUser

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

type AuthResult struct {
	AccessToken string       `json:"access_token"`
	User        *UserPayload `json:"user"`
}

// RegistrationResult contains only the created user payload (no token)
type RegistrationResult struct {
	User *UserPayload `json:"user"`
}

type UserPayload struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*RegistrationResult, error)
	Login(ctx context.Context, input LoginInput) (*AuthResult, error)
}

type authService struct {
	repo   repo.UserRepository
	tokens *jwtx.TokenManager
}

func NewAuthService(repo repo.UserRepository, tokens *jwtx.TokenManager) AuthService {
	return &authService{repo: repo, tokens: tokens}
}

func (svc *authService) Register(ctx context.Context, input RegisterInput) (*RegistrationResult, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" || strings.TrimSpace(input.Password) == "" || strings.TrimSpace(input.FullName) == "" {
		return nil, ErrInvalidCredentials
	}

	if _, err := svc.repo.FindByEmail(ctx, email); err == nil {
		return nil, ErrEmailAlreadyUsed
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashed, err := bcryptx.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	role := strings.TrimSpace(strings.ToUpper(input.Role))
	if _, ok := allowedRoles[role]; !ok {
		role = defaultRole
	}

	user := &entity.User{
		FullName:       strings.TrimSpace(input.FullName),
		Email:          email,
		HashedPassword: hashed,
		Role:           role,
	}

	if err := svc.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &RegistrationResult{
		User: &UserPayload{
			ID:       user.ID.String(),
			FullName: user.FullName,
			Email:    user.Email,
		},
	}, nil
}

func (svc *authService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" || strings.TrimSpace(input.Password) == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := svc.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcryptx.CompareHash(user.HashedPassword, input.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := svc.tokens.SignToken(user.ID.String(), user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return buildAuthResult(user, token), nil
}

func buildAuthResult(user *entity.User, token string) *AuthResult {
	return &AuthResult{
		AccessToken: token,
		User: &UserPayload{
			ID:       user.ID.String(),
			FullName: user.FullName,
			Email:    user.Email,
		},
	}
}
