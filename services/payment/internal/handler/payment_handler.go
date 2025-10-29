package handler

import (
	"net/http"
	"payment/internal/service"
	"pkg/httpx"

	"github.com/gin-gonic/gin"
)

type Handler struct{ svc *service.Service }

func NewHandler(s *service.Service) *Handler { return &Handler{svc: s} }

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

func (h *Handler) BindRoutes(r *gin.Engine) {
	r.POST("/bookings/:id/pay", h.CreatePayment)
	r.POST("/payments/midtrans/webhook", h.Webhook)
	r.POST("/payments/:id/refund", h.Refund)
}
