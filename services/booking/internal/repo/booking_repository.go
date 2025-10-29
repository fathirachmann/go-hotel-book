package repo

import (
	"booking/internal/entity"
	"context"

	"gorm.io/gorm"
)

type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(ctx context.Context, b *entity.Booking) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Detach items to avoid GORM auto-saving associations
		items := b.Items
		b.Items = nil
		if err := tx.Create(b).Error; err != nil {
			return err
		}
		// Ensure BookingID is set on items (if not set by caller)
		if len(items) > 0 {
			for i := range items {
				items[i].BookingID = b.ID
			}
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
			// attach created items back to booking for response payloads
			b.Items = items
		}
		return nil
	})
}

func (r *BookingRepository) ListByUser(ctx context.Context, userID string) ([]entity.Booking, error) {
	var list []entity.Booking
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, id string, status entity.Status) error {
	return r.db.WithContext(ctx).Model(&entity.Booking{}).Where("id = ?", id).Update("status", status).Error
}

func (r *BookingRepository) GetByID(ctx context.Context, id string) (*entity.Booking, error) {
	var b entity.Booking
	if err := r.db.WithContext(ctx).
		Preload("Items").
		First(&b, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}
