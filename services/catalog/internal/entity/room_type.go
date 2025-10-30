package entity

import (
	"time"
)

// RoomType represents a sellable room configuration within the hotel.
type RoomType struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:120;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	BasePrice   int64     `gorm:"not null" json:"base_price"`
	Capacity    int       `gorm:"not null" json:"capacity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
