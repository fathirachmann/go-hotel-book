package handler

import (
	"booking/internal/entity"
	"booking/internal/service"
	"errors"
	"io"
	"net/http"
	"time"

	"pkg/httpx"
	"pkg/jwtx"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	svc *service.Service
	tm  *jwtx.TokenManager
}

func NewHandler(s *service.Service, tm *jwtx.TokenManager) *Handler {
	return &Handler{svc: s, tm: tm}
}

type CreateRequest struct {
	UserID   string                     `json:"user_id" binding:"required"`
	CheckIn  time.Time                  `json:"check_in" binding:"required"`
	CheckOut time.Time                  `json:"check_out" binding:"required"`
	Guests   int                        `json:"guests"`
	Items    []entity.CreateBookingItem `json:"items" binding:"required,min=1,dive"`
	Email    string                     `json:"email" binding:"required,email"`
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
		UserID:   req.UserID,
		CheckIn:  req.CheckIn,
		CheckOut: req.CheckOut,
		Guests:   req.Guests,
		Items:    req.Items,
		Email:    req.Email,
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

// getUserID extracts user id from Authorization bearer token.
func (h *Handler) getUserID(c *gin.Context) (string, error) {
	if h.tm == nil {
		return "", errors.New("auth not configured")
	}
	tok, err := jwtx.ExtractToken(c.Request)
	if err != nil {
		return "", err
	}
	claims, err := h.tm.VerifyToken(tok)
	if err != nil {
		return "", err
	}
	if claims == nil || claims.UserID == "" {
		return "", errors.New("invalid token claims")
	}
	return claims.UserID, nil
}

// GetMyBookings lists bookings owned by the logged-in user.
func (h *Handler) GetMyBookings(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, httpx.ErrorResponse{Error: "unauthorized"})
		return
	}
	list, err := h.svc.ListMine(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, httpx.OK(list))
}

// GetBookingDetail returns a user's booking detail by ID.
func (h *Handler) GetBookingDetail(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, httpx.ErrorResponse{Error: "unauthorized"})
		return
	}
	id := c.Param("id")
	booking, err := h.svc.GetMineByID(c.Request.Context(), id, userID)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, httpx.ErrorResponse{Error: "booking not found"})
		case err.Error() == "forbidden":
			c.JSON(http.StatusForbidden, httpx.ErrorResponse{Error: "forbidden"})
		default:
			c.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, httpx.OK(booking))
}
