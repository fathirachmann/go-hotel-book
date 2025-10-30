package repo

import (
	"catalog/internal/entity"
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InventoryRepository exposes per-day stock persistence.
type InventoryRepository interface {
	Upsert(ctx context.Context, inv *entity.RoomInventory) error
	MinAvailable(ctx context.Context, roomTypeID uint, from, to time.Time) (int, error)
	Prices(ctx context.Context, roomTypeID uint, from, to time.Time) ([]int64, error)
	DeleteAll(ctx context.Context) error
}

type inventoryRepository struct {
	db *gorm.DB
}

// NewInventoryRepository provides a GORM-backed inventory repository.
func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) Upsert(ctx context.Context, inv *entity.RoomInventory) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "room_type_id"}, {Name: "inv_date"}},
			DoUpdates: clause.AssignmentColumns([]string{"total_rooms", "available_rooms", "price_override"}),
		}).
		Create(inv).Error
}

func (r *inventoryRepository) MinAvailable(ctx context.Context, roomTypeID uint, from, to time.Time) (int, error) {
	var minAvail *int
	if err := r.db.WithContext(ctx).
		Model(&entity.RoomInventory{}).
		Select("MIN(available_rooms)").
		Where("room_type_id = ? AND inv_date >= ? AND inv_date < ?", roomTypeID, from, to).
		Scan(&minAvail).Error; err != nil {
		return 0, err
	}
	if minAvail == nil {
		return 0, nil
	}
	return *minAvail, nil
}

func (r *inventoryRepository) Prices(ctx context.Context, roomTypeID uint, from, to time.Time) ([]int64, error) {
	var rows []entity.RoomInventory
	if err := r.db.WithContext(ctx).
		Where("room_type_id = ? AND inv_date >= ? AND inv_date < ?", roomTypeID, from, to).
		Order("inv_date ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	prices := make([]int64, len(rows))
	for i, row := range rows {
		if row.PriceOverride != nil {
			prices[i] = *row.PriceOverride
			continue
		}
		prices[i] = -1
	}
	return prices, nil
}

func (r *inventoryRepository) DeleteAll(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("1 = 1").Delete(&entity.RoomInventory{}).Error
}
