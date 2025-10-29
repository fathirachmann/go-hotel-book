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
	CheckIn  time.Time                  `json:"check_in" binding:"required"`
	CheckOut time.Time                  `json:"check_out" binding:"required"`
	Guests   int                        `json:"guests"`
	FullName string                     `json:"full_name" binding:"required"`
	Items    []entity.CreateBookingItem `json:"items" binding:"required,min=1,dive"`
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
	// derive user from token
	userID, err := h.requireUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, httpx.ErrorResponse{Error: "unauthorized"})
		return
	}
	email := h.userEmail(c)
	b, err := h.svc.Create(c.Request.Context(), entity.CreateBookingInput{
		UserID:   userID,
		CheckIn:  req.CheckIn,
		CheckOut: req.CheckOut,
		Guests:   req.Guests,
		FullName: req.FullName,
		Email:    email,
		Items:    req.Items,
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

// PostInternalUpdateStatus updates a booking status via internal system calls (e.g., from Payment service).
type internalStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *Handler) PostInternalUpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req internalStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		return
	}
	// Map to booking entity.Status
	var st entity.Status
	switch req.Status {
	case "PAID":
		st = entity.StatusPaid
	case "REFUNDED":
		st = entity.StatusRefunded
	case "CANCELLED", "EXPIRED":
		st = entity.StatusCancelled
	default:
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: "invalid status"})
		return
	}
	if err := h.svc.RepoUpdateStatus(c.Request.Context(), id, st); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, httpx.ErrorResponse{Error: "booking not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, httpx.OK(gin.H{"status": st}))
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

// authMiddleware ensures a valid bearer token and sets user_id in context.
func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := h.getClaims(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httpx.ErrorResponse{Error: "unauthorized"})
			return
		}
		c.Set("user_id", claims.UserID)
		if claims.Email != "" {
			c.Set("email", claims.Email)
		}
		c.Next()
	}
}

// getClaims extracts full claims from Authorization bearer token.
func (h *Handler) getClaims(c *gin.Context) (*jwtx.AccessClaims, error) {
	if h.tm == nil {
		return nil, errors.New("auth not configured")
	}
	tok, err := jwtx.ExtractToken(c.Request)
	if err != nil {
		return nil, err
	}
	claims, err := h.tm.VerifyToken(tok)
	if err != nil {
		return nil, err
	}
	if claims == nil || claims.UserID == "" {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

// requireUser retrieves user_id from context (populated by authMiddleware) or parses token as fallback.
func (h *Handler) requireUser(c *gin.Context) (string, error) {
	if v, ok := c.Get("user_id"); ok {
		if s, ok := v.(string); ok && s != "" {
			return s, nil
		}
	}
	claims, err := h.getClaims(c)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// userEmail returns email from context or token claims.
func (h *Handler) userEmail(c *gin.Context) string {
	if v, ok := c.Get("email"); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	if claims, err := h.getClaims(c); err == nil && claims.Email != "" {
		return claims.Email
	}
	return ""
}

// GetMyBookings lists bookings owned by the logged-in user.
func (h *Handler) GetMyBookings(c *gin.Context) {
	userID, err := h.requireUser(c)
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
	userID, err := h.requireUser(c)
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

// DeleteBooking removes a user's booking by ID.
func (h *Handler) DeleteBooking(c *gin.Context) {
	userID, err := h.requireUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, httpx.ErrorResponse{Error: "unauthorized"})
		return
	}
	id := c.Param("id")
	if err := h.svc.DeleteMine(c.Request.Context(), id, userID); err != nil {
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
	c.Status(http.StatusNoContent)
}
