package repo

import (
	"context"
	"encoding/json"
	"time"

	"payment/internal/entity"

	"gorm.io/gorm"
)

// PaymentRepository defines persistence operations for payments.
type PaymentRepository interface {
	Create(ctx context.Context, payment *entity.Payment) error
	UpdateStatus(ctx context.Context, orderID string, status entity.Status, payload []byte, paidAt *time.Time) error
	UpdateSnapAttributes(ctx context.Context, orderID, token, redirectURL string) error
	GetByOrderID(ctx context.Context, orderID string) (*entity.Payment, error)
}

type paymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository returns a GORM-backed payment repository.
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, payment *entity.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, orderID string, status entity.Status, payload []byte, paidAt *time.Time) error {
	updates := map[string]any{
		"status":            status,
		"last_notification": json.RawMessage(payload),
	}
	if paidAt != nil {
		updates["paid_at"] = paidAt
	}
	res := r.db.WithContext(ctx).
		Model(&entity.Payment{}).
		Where("order_id = ?", orderID).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *paymentRepository) UpdateSnapAttributes(ctx context.Context, orderID, token, redirectURL string) error {
	res := r.db.WithContext(ctx).
		Model(&entity.Payment{}).
		Where("order_id = ?", orderID).
		Updates(map[string]any{
			"snap_token":   token,
			"redirect_url": redirectURL,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *paymentRepository) GetByOrderID(ctx context.Context, orderID string) (*entity.Payment, error) {
	var payment entity.Payment
	if err := r.db.WithContext(ctx).First(&payment, "order_id = ?", orderID).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}
