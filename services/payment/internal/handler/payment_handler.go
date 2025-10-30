package handler

import (
	"net/http"
	"payment/internal/service"
	"pkg/httpx"
	"pkg/jwtx"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.Service
	tm  *jwtx.TokenManager
}

func NewHandler(s *service.Service, tm *jwtx.TokenManager) *Handler { return &Handler{svc: s, tm: tm} }

type payRequest struct {
	Amount int64 `json:"amount" binding:"required,gt=0"`
}

func (h *Handler) CreatePayment(c *gin.Context) {
	bookingID := c.Param("id")
	var req payRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		return
	}
	_, resp, err := h.svc.CreatePayment(c.Request.Context(), bookingID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, httpx.OK(resp))
}

type createPaymentBody struct {
	BookingID     string `json:"booking_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	CustomerEmail string `json:"customer_email"`
	CustomerName  string `json:"customer_name"`
}

// CreatePaymentBody accepts POST /payments with JSON body and creates payment
func (h *Handler) CreatePaymentBody(c *gin.Context) {
	var req createPaymentBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		return
	}
	_, resp, err := h.svc.CreatePayment(c.Request.Context(), req.BookingID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, httpx.OK(resp))
}

type webhookPayload struct {
	OrderID           string `json:"order_id"`
	TransactionStatus string `json:"transaction_status"`
	GrossAmount       string `json:"gross_amount"`
	TransactionID     string `json:"transaction_id"`
	SignatureKey      string `json:"signature_key"`
}

func (h *Handler) Webhook(c *gin.Context) {
	var payload webhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		return
	}
	if err := h.svc.HandleMidtransWebhook(c.Request.Context(), service.MidtransWebhookPayload(payload)); err != nil {
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type refundRequest struct {
	Amount int64 `json:"amount"`
}

func (h *Handler) Refund(c *gin.Context) {
	paymentID := c.Param("id")
	var req refundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Amount = 0
	}
	if err := h.svc.Refund(c.Request.Context(), paymentID, req.Amount); err != nil {
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, httpx.OK(gin.H{"status": "REFUNDED"}))
}

// authMiddleware verifies JWT and injects claims into context
func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := jwtx.ExtractToken(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httpx.ErrorResponse{Error: err.Error()})
			return
		}
		claims, err := h.tm.VerifyToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httpx.ErrorResponse{Error: err.Error()})
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}

func (h *Handler) getClaims(c *gin.Context) *jwtx.AccessClaims {
	v, ok := c.Get("claims")
	if !ok {
		return nil
	}
	cl, _ := v.(*jwtx.AccessClaims)
	return cl
}

// GetPayments lists payments for the authenticated user
func (h *Handler) GetPayments(c *gin.Context) {
	claims := h.getClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, httpx.ErrorResponse{Error: "missing claims"})
		return
	}
	items, err := h.svc.ListByUserID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, httpx.OK(items))
}

func (h *Handler) BindRoutes(r *gin.Engine) {
	// Public webhook
	r.POST("/payments/midtrans/webhook", h.Webhook)

	// Authenticated routes
	auth := r.Group("")
	auth.Use(h.authMiddleware())
	auth.POST("/bookings/:id/pay", h.CreatePayment)
	auth.POST("/payments/:id/refund", h.Refund)
	auth.GET("/payments", h.GetPayments)
}
