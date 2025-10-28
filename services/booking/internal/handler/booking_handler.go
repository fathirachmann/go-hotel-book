package handler

import (
	"booking/internal/entity"
	"booking/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(s *service.Service) *Handler {
	return &Handler{svc: s}
}

type CreateRequest struct {
	UserID       string               `json:"user_id" boinding:"required"`
	CheckInDate  time.Time            `json:"check_in_date" binding:"required"`
	CheckOutDate time.Time            `json:"check_out_date" binding:"required"`
	Items        []entity.BookingItem `json:"items" binding:"required"`
	Email        string               `json:"email" binding:"required,email,unique"`
}

func (h *Handler) PostBooking(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := h.svc.Create(c.Request.Context(), entity.CreateBookingInput{
		UserID:       req.UserID,
		CheckInDate:  req.CheckInDate,
		CheckOutDate: req.CheckOutDate,
		Items:        req.Items,
		Email:        req.Email,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"booking": b,
	})
}
