package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoomType represents a sellable room configuration within the hotel.
type RoomType struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"size:120;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	BasePrice   int64     `gorm:"not null" json:"base_price"`
	Capacity    int       `gorm:"not null" json:"capacity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BeforeCreate ensures a UUID is present for new room types.
func (rt *RoomType) BeforeCreate(_ *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}
