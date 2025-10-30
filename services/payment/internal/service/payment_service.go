package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"payment/internal/entity"
)

type Service struct {
	payRepo entity.PaymentRepo
	refRepo entity.RefundRepo
	book    entity.BookingClient
}

func NewPaymentService(p entity.PaymentRepo, r entity.RefundRepo, b entity.BookingClient) *Service {
	return &Service{payRepo: p, refRepo: r, book: b}
}

type CreatePaymentResponse struct {
	SnapToken   string `json:"snap_token"`
	RedirectURL string `json:"redirect_url"`
	Amount      int64  `json:"amount"`
}

func (s *Service) CreatePayment(ctx context.Context, bookingID string, amount int64) (*entity.Payment, *CreatePaymentResponse, error) {
	orderID := fmt.Sprintf("BO-%s", bookingID)
	p := &entity.Payment{
		BookingID: bookingID,
		OrderID:   orderID,
		Amount:    amount,
		Provider:  "midtrans",
		Status:    entity.PayPending,
	}
	if err := s.payRepo.Create(ctx, p); err != nil {
		return nil, nil, err
	}
	resp := &CreatePaymentResponse{
		SnapToken:   "mock-" + orderID,
		RedirectURL: "https://mock.midtrans/redirect/" + orderID,
		Amount:      amount,
	}
	return p, resp, nil
}

type MidtransWebhookPayload struct {
	OrderID           string `json:"order_id"`
	TransactionStatus string `json:"transaction_status"`
	GrossAmount       string `json:"gross_amount"`
	TransactionID     string `json:"transaction_id"`
	SignatureKey      string `json:"signature_key"`
}

func (s *Service) HandleMidtransWebhook(ctx context.Context, payload MidtransWebhookPayload) error {
	if payload.OrderID == "" {
		return errors.New("missing order_id")
	}
	pay, err := s.payRepo.FindByOrderID(ctx, payload.OrderID)
	if err != nil {
		return err
	}
	rawBytes, _ := json.Marshal(payload)
	switch payload.TransactionStatus {
	case "settlement":
		if err := s.payRepo.UpdateStatus(ctx, pay.ID, entity.PaySettlement, string(rawBytes), payload.TransactionID); err != nil {
			return err
		}
		return s.book.UpdateStatusPaid(ctx, pay.BookingID)
	case "expire", "cancel", "deny":
		if err := s.payRepo.UpdateStatus(ctx, pay.ID, entity.PayExpire, string(rawBytes), payload.TransactionID); err != nil {
			return err
		}
		return s.book.UpdateStatusExpired(ctx, pay.BookingID)
	default:
		// ignore other statuses for mock
		return nil
	}
}

func (s *Service) Refund(ctx context.Context, paymentID string, amount int64) error {
	if paymentID == "" {
		return errors.New("missing payment id")
	}
	// Persist refund
	rf := &entity.Refund{PaymentID: paymentID, Amount: amount, Status: "SUCCESS"}
	if err := s.refRepo.Create(ctx, rf); err != nil {
		return err
	}
	// Update payment and notify booking
	if err := s.payRepo.UpdateStatus(ctx, paymentID, entity.PayRefunded, "{}", ""); err != nil {
		return err
	}
	// We need bookingID; in this simplified mock, caller should already know and webhook would have updated
	// Since UpdateStatus uses paymentID only, we cannot infer bookingID here without extra fetch; skip notify if unknown
	return nil
}

// ListByUserID returns all payments for bookings owned by the given user ID.
func (s *Service) ListByUserID(ctx context.Context, userID string) ([]entity.Payment, error) {
	return s.payRepo.ListByUserID(ctx, userID)
}
