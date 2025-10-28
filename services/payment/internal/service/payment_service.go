package service

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"payment/internal/entity"
	"payment/internal/repo"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"gorm.io/gorm"
)

var (
	// ErrPaymentNotFound indicates we could not locate the payment row referenced by provider callbacks.
	ErrPaymentNotFound = errors.New("payment not found")
	// ErrInvalidSignature indicates provider webhook failed signature validation.
	ErrInvalidSignature = errors.New("invalid midtrans signature")
)

// SnapGateway abstracts Midtrans Snap client for easier testing.
type SnapGateway interface {
	CreateTransaction(req *snap.Request) (*snap.Response, error)
}

// PaymentService orchestrates payment requests and asynchronous updates.
type PaymentService struct {
	repo      repo.PaymentRepository
	gateway   SnapGateway
	serverKey string
}

// NewPaymentService constructs a PaymentService instance.
func NewPaymentService(repo repo.PaymentRepository, gateway SnapGateway, serverKey string) *PaymentService {
	return &PaymentService{repo: repo, gateway: gateway, serverKey: serverKey}
}

// RequestPaymentInput represents payload required to start a payment.
type RequestPaymentInput struct {
	BookingID     string
	Amount        int64
	CustomerEmail string
	CustomerName  string
}

// RequestPayment triggers a new payment request through Midtrans Snap.
func (s *PaymentService) RequestPayment(ctx context.Context, in RequestPaymentInput) (*entity.Payment, error) {
	orderID := buildOrderID(in.BookingID)
	payment := &entity.Payment{
		ID:            uuid.New(),
		BookingID:     in.BookingID,
		OrderID:       orderID,
		Provider:      "midtrans",
		Amount:        in.Amount,
		Currency:      "IDR",
		Status:        entity.StatusPending,
		CustomerEmail: in.CustomerEmail,
		CustomerName:  in.CustomerName,
	}

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: in.Amount,
		},
		CustomerDetail: &midtrans.CustomerDetails{
			Email: in.CustomerEmail,
			FName: in.CustomerName,
		},
		Items: &[]midtrans.ItemDetails{
			{
				ID:    in.BookingID,
				Name:  fmt.Sprintf("Booking %s", in.BookingID),
				Price: in.Amount,
				Qty:   1,
			},
		},
	}

	resp, err := s.gateway.CreateTransaction(req)
	if err != nil {
		return nil, err
	}

	payment.RedirectURL = resp.RedirectURL
	payment.SnapToken = resp.Token

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateSnapAttributes(ctx, orderID, resp.Token, resp.RedirectURL); err != nil {
		return nil, err
	}

	return payment, nil
}

// MidtransNotification reflects subset of fields sent by Midtrans webhook.
type MidtransNotification struct {
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
	TransactionTime   string `json:"transaction_time"`
}

// ProcessMidtransNotification validates webhook and updates persisted payment state.
func (s *PaymentService) ProcessMidtransNotification(ctx context.Context, body []byte) (*entity.Payment, error) {
	var notif MidtransNotification
	if err := json.Unmarshal(body, &notif); err != nil {
		return nil, err
	}

	if notif.OrderID == "" {
		return nil, errors.New("missing order_id in notification")
	}

	if s.serverKey != "" && notif.SignatureKey != "" {
		if err := s.verifySignature(notif); err != nil {
			return nil, err
		}
	}

	status := mapMidtransStatus(notif.TransactionStatus, notif.FraudStatus)

	var paidAt *time.Time
	if status == entity.StatusPaid && notif.TransactionTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", notif.TransactionTime); err == nil {
			paidAt = &t
		}
	}

	if err := s.repo.UpdateStatus(ctx, notif.OrderID, status, body, paidAt); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	updated, err := s.repo.GetByOrderID(ctx, notif.OrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	return updated, nil
}

// MockUpdateStatus allows manual status override (useful for local tests without callbacks).
func (s *PaymentService) MockUpdateStatus(ctx context.Context, orderID string, status entity.Status) (*entity.Payment, error) {
	if orderID == "" {
		return nil, errors.New("orderID cannot be empty")
	}

	var paidAt *time.Time
	if status == entity.StatusPaid {
		now := time.Now().UTC()
		paidAt = &now
	}

	if err := s.repo.UpdateStatus(ctx, orderID, status, nil, paidAt); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	return s.repo.GetByOrderID(ctx, orderID)
}

// GetByOrderID retrieves payment by orderID.
func (s *PaymentService) GetByOrderID(ctx context.Context, orderID string) (*entity.Payment, error) {
	payment, err := s.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return payment, nil
}

func (s *PaymentService) verifySignature(n MidtransNotification) error {
	data := n.OrderID + n.StatusCode + n.GrossAmount + s.serverKey
	sum := sha512.Sum512([]byte(data))
	if hex.EncodeToString(sum[:]) != strings.ToLower(n.SignatureKey) {
		return ErrInvalidSignature
	}
	return nil
}

func buildOrderID(bookingID string) string {
	base := strings.ReplaceAll(strings.TrimSpace(bookingID), " ", "")
	if base == "" {
		base = uuid.NewString()
	}
	timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	return fmt.Sprintf("%s-%s", strings.ToLower(base), timestamp)
}

func mapMidtransStatus(transactionStatus, fraudStatus string) entity.Status {
	ts := strings.ToLower(transactionStatus)
	fs := strings.ToLower(fraudStatus)

	switch ts {
	case "capture":
		if fs == "challenge" {
			return entity.StatusPending
		}
		return entity.StatusPaid
	case "settlement":
		return entity.StatusPaid
	case "pending":
		return entity.StatusPending
	case "deny", "cancel":
		return entity.StatusFailed
	case "expire":
		return entity.StatusExpired
	default:
		return entity.StatusPending
	}
}

// NewSnapClient creates a production-ready Midtrans Snap client.
func NewSnapClient(serverKey string, env midtrans.EnvironmentType) SnapGateway {
	var client snap.Client
	client.New(serverKey, env)
	return &snapClientAdapter{client: client}
}

type snapClientAdapter struct {
	client snap.Client
}

func (s *snapClientAdapter) CreateTransaction(req *snap.Request) (*snap.Response, error) {
	resp, err := s.client.CreateTransaction(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// MockSnapClient returns fake snap responses for local development.
type MockSnapClient struct{}

func (MockSnapClient) CreateTransaction(req *snap.Request) (*snap.Response, error) {
	token := uuid.NewString()
	return &snap.Response{
		Token:       token,
		RedirectURL: fmt.Sprintf("https://mock-payments.local/redirect/%s", token),
	}, nil
}
