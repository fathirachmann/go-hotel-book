package repo

import (
	"context"
	"payment/internal/entity"

	"gorm.io/gorm"
)

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *paymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, p *entity.Payment) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *paymentRepository) FindByOrderID(ctx context.Context, orderID string) (*entity.Payment, error) {
	var p entity.Payment
	if err := r.db.WithContext(ctx).First(&p, "order_id = ?", orderID).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, id string, status entity.PaymentStatus, raw string, providerRef string) error {
	res := r.db.WithContext(ctx).
		Model(&entity.Payment{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       status,
			"raw_payload":  raw,
			"provider_ref": providerRef,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListByUserID returns payments joined with booking.bookings filtered by user ID
func (r *paymentRepository) ListByUserID(ctx context.Context, userID string) ([]entity.Payment, error) {
	var res []entity.Payment
	// Use explicit schema qualification to avoid search_path issues
	q := `SELECT p.*
		  FROM "payment"."payments" p
		  JOIN "booking"."bookings" b ON b.id::text = p.booking_id
		  WHERE b.user_id::text = ?
		  ORDER BY p.created_at DESC`
	if err := r.db.WithContext(ctx).Raw(q, userID).Scan(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

// GetBookingTotal returns the total price for a booking from the booking schema
func (r *paymentRepository) GetBookingTotal(ctx context.Context, bookingID string) (int64, error) {
	var total int64
	q := `SELECT b.total FROM "booking"."bookings" b WHERE b.id::text = ? LIMIT 1`
	if err := r.db.WithContext(ctx).Raw(q, bookingID).Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
