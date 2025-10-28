package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoomInventory captures day-level availability and pricing overrides.
type RoomInventory struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RoomTypeID     uuid.UUID `gorm:"type:uuid;index;not null" json:"room_type_id"`
	InvDate        time.Time `gorm:"type:date;index;not null" json:"inv_date"`
	TotalRooms     int       `gorm:"not null" json:"total_rooms"`
	AvailableRooms int       `gorm:"not null" json:"available_rooms"`
	PriceOverride  *int64    `json:"price_override"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// BeforeCreate assigns a UUID when the inventory row is inserted.
func (ri *RoomInventory) BeforeCreate(_ *gorm.DB) error {
	if ri.ID == uuid.Nil {
		ri.ID = uuid.New()
	}
	return nil
}
