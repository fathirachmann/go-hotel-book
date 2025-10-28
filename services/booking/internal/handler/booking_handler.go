package handler

import (
	"booking/internal/entity"
	"booking/internal/service"
	"errors"
	"io"
	"net/http"
	"time"

	"pkg/httpx"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(s *service.Service) *Handler {
	return &Handler{svc: s}
}

type CreateRequest struct {
	UserID       string               `json:"user_id" binding:"required"`
	CheckInDate  time.Time            `json:"check_in_date" binding:"required"`
	CheckOutDate time.Time            `json:"check_out_date" binding:"required"`
	Items        []entity.BookingItem `json:"items" binding:"required"`
	Email        string               `json:"email" binding:"required,email"`
}

type refundRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) PostBooking(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
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
		c.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, httpx.OK(b))
}

// PostCheckIn marks booking as checked-in once payment is confirmed.
func (h *Handler) PostCheckIn(c *gin.Context) {
	bookingID := c.Param("id")
	booking, err := h.svc.CheckIn(c.Request.Context(), bookingID)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, httpx.ErrorResponse{Error: "booking not found"})
		case errors.Is(err, service.ErrBookingNotPaid), errors.Is(err, service.ErrBookingAlreadyHandled):
			c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, httpx.OK(booking))
}

// PostRefund cancels booking and requests payment refund when eligible.
func (h *Handler) PostRefund(c *gin.Context) {
	bookingID := c.Param("id")
	var req refundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if errors.Is(err, io.EOF) {
			// empty body allowed -> use default reason
		} else {
			c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
			return
		}
	}

	booking, err := h.svc.Refund(c.Request.Context(), bookingID, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, httpx.ErrorResponse{Error: "booking not found"})
		case errors.Is(err, service.ErrBookingNotPaid), errors.Is(err, service.ErrBookingAlreadyHandled):
			c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, httpx.OK(booking))
}
