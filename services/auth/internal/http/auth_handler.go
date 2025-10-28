package http

import (
	"errors"
	"net/http"

	"auth/internal/usecase"

	"pkg/bcryptx"
	"pkg/httpx"

	"github.com/gin-gonic/gin"
)

// AuthHandler wires HTTP routes to the auth usecase.
type AuthHandler struct {
	uc usecase.AuthUsecase
}

// NewAuthHandler constructs an AuthHandler instance.
func NewAuthHandler(uc usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) UserRoutes(router *gin.Engine) {
	authGroup := router.Group("/api/v1/auth")
	authGroup.POST("/register", h.handleRegister)
	authGroup.POST("/login", h.handleLogin)
}

type registerRequest struct {
	FullName string `json:"full_name" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *AuthHandler) handleRegister(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		return
	}

	hashedPassword, err := bcryptx.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: "failed to hash password"})
		return
	}

	result, err := h.uc.Register(c.Request.Context(), usecase.RegisterInput{
		FullName: req.FullName,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     req.Role,
	})

	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, httpx.OK(result))
}

func (h *AuthHandler) handleLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.uc.Login(c.Request.Context(), usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, httpx.OK(result))
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrEmailAlreadyUsed):
		c.JSON(http.StatusConflict, httpx.ErrorResponse{Error: err.Error()})
	case errors.Is(err, usecase.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, httpx.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: "internal server error"})
	}
}
