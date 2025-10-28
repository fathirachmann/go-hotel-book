package handler

import (
	"bytes"
	"io"
	"net/http"

	"payment/internal/entity"
	"payment/internal/service"

	"pkg/httpx"

	"github.com/gin-gonic/gin"
)

// Handler wires HTTP routes to payment service.
type Handler struct {
	svc *service.PaymentService
}

// NewHandler creates a payment HTTP handler set.
func NewHandler(svc *service.PaymentService) *Handler {
	return &Handler{svc: svc}
}

type createPaymentRequest struct {
	BookingID     string `json:"booking_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	CustomerEmail string `json:"customer_email" binding:"required,email"`
	CustomerName  string `json:"customer_name"`
}

type paymentResponse struct {
	OrderID     string `json:"order_id"`
	BookingID   string `json:"booking_id"`
	Amount      int64  `json:"amount"`
	Status      string `json:"status"`
	RedirectURL string `json:"redirect_url"`
	SnapToken   string `json:"snap_token"`
}

func toPaymentResponse(p *entity.Payment) paymentResponse {
	return paymentResponse{
		OrderID:     p.OrderID,
		BookingID:   p.BookingID,
		Amount:      p.Amount,
		Status:      string(p.Status),
		RedirectURL: p.RedirectURL,
		SnapToken:   p.SnapToken,
	}
}

// PostPayment handles requests to initiate payment collection.
func (h *Handler) PostPayment(c *gin.Context) {
	var req createPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		return
	}

	payment, err := h.svc.RequestPayment(c.Request.Context(), service.RequestPaymentInput{
		BookingID:     req.BookingID,
		Amount:        req.Amount,
		CustomerEmail: req.CustomerEmail,
		CustomerName:  req.CustomerName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, httpx.OK(toPaymentResponse(payment)))
}

// GetPayment fetches payment detail for given orderID.
func (h *Handler) GetPayment(c *gin.Context) {
	orderID := c.Param("orderID")
	payment, err := h.svc.GetByOrderID(c.Request.Context(), orderID)
	if err != nil {
		switch err {
		case service.ErrPaymentNotFound:
			c.JSON(http.StatusNotFound, httpx.ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, httpx.OK(toPaymentResponse(payment)))
}

// MidtransWebhook processes asynchronous payment notifications.
func (h *Handler) MidtransWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: "failed to read request body"})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	payment, err := h.svc.ProcessMidtransNotification(c.Request.Context(), body)
	if err != nil {
		switch err {
		case service.ErrInvalidSignature:
			c.JSON(http.StatusUnauthorized, httpx.ErrorResponse{Error: err.Error()})
		case service.ErrPaymentNotFound:
			c.JSON(http.StatusNotFound, httpx.ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, httpx.OK(toPaymentResponse(payment)))
}

type mockStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// MockStatus is convenient endpoint to simulate provider callbacks locally.
func (h *Handler) MockStatus(c *gin.Context) {
	orderID := c.Param("orderID")
	var req mockStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		return
	}

	status, err := entity.ParseStatus(req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
		return
	}

	payment, err := h.svc.MockUpdateStatus(c.Request.Context(), orderID, status)
	if err != nil {
		switch err {
		case service.ErrPaymentNotFound:
			c.JSON(http.StatusNotFound, httpx.ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, httpx.OK(toPaymentResponse(payment)))
}

// BindRoutes attaches payment routes to gin engine.
func (h *Handler) BindRoutes(r *gin.Engine) {
	r.POST("/payments", h.PostPayment)
	r.GET("/payments/:orderID", h.GetPayment)
	r.POST("/payments/mock/:orderID", h.MockStatus)
	r.POST("/payments/webhook/midtrans", h.MidtransWebhook)
}
