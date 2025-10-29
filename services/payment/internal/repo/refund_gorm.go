package repo

import (
	"context"
	"payment/internal/entity"

	"gorm.io/gorm"
)

type refundRepository struct {
	db *gorm.DB
}

func NewRefundRepository(db *gorm.DB) *refundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) Create(ctx context.Context, rf *entity.Refund) error {
	return r.db.WithContext(ctx).Create(rf).Error
}
