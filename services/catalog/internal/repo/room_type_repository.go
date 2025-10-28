package repo

import (
	"catalog/internal/entity"
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RoomTypeRepository exposes persistence operations for room types.
type RoomTypeRepository interface {
	List(ctx context.Context) ([]entity.RoomType, error)
	GetByIDs(ctx context.Context, ids []string) ([]entity.RoomType, error)
	Upsert(ctx context.Context, roomType *entity.RoomType) error
}

type roomTypeRepository struct {
	db *gorm.DB
}

// NewRoomTypeRepository provides a GORM-backed room type repository.
func NewRoomTypeRepository(db *gorm.DB) RoomTypeRepository {
	return &roomTypeRepository{db: db}
}

func (r *roomTypeRepository) List(ctx context.Context) ([]entity.RoomType, error) {
	var out []entity.RoomType
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *roomTypeRepository) GetByIDs(ctx context.Context, ids []string) ([]entity.RoomType, error) {
	var out []entity.RoomType
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *roomTypeRepository) Upsert(ctx context.Context, roomType *entity.RoomType) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "description", "base_price", "capacity"}),
		}).
		Create(roomType).Error
}
