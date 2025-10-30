package handler

import (
	"errors"
	"net/http"

	"auth/internal/service"

	"pkg/httpx"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type registerRequest struct {
	FullName string `json:"full_name" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *AuthHandler) HandleRegister(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.svc.Register(c.Request.Context(), service.RegisterInput{
		FullName: req.FullName,
		Email:    req.Email,
		Password: req.Password,
		Role:     "USER",
	})

	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created successfully",
		"data":    result,
	})
}

func (h *AuthHandler) HandleLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.svc.Login(c.Request.Context(), service.LoginInput{
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
	case errors.Is(err, service.ErrEmailAlreadyUsed):
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, httpx.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: "internal server error"})
	}
}
