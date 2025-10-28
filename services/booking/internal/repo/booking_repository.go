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
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, id string, status entity.Status) error {
	return r.db.WithContext(ctx).Model(&entity.Booking{}).Where("id = ?", id).Update("status", status).Error
}

func (r *BookingRepository) GetByID(ctx context.Context, id string) (*entity.Booking, error) {
	var b entity.Booking

	if err := r.db.WithContext(ctx).First(&b, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}
