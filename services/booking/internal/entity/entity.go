package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Status string

const (
	StatusUnpaid     Status = "UNPAID"
	StatusPaid       Status = "PAID"
	StatusCancelled  Status = "CANCELLED"
	StatusCheckedIn  Status = "CHECKED_IN"
	StatusCheckedOut Status = "CHECKED_OUT"
	StatusRefunded   Status = "REFUNDED"
)

type Booking struct {
	ID           string        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       string        `gorm:"index" json:"user_id"`
	Code         string        `gorm:"uniqueIndex" json:"code"`
	CheckInDate  time.Time     `json:"check_in_date"`
	CheckOutDate time.Time     `json:"check_out_date"`
	Nights       int           `json:"nights"`
	Guests       int           `json:"guests"`
	Subtotal     int64         `json:"subtotal"`
	Taxes        int64         `json:"taxes"`
	Total        int64         `json:"total"`
	Status       Status        `gorm:"index" json:"status"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Items        []BookingItem `gorm:"foreignKey:BookingID" json:"items"`
}

func (b *Booking) BeforeCreate(_ *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	if b.Code == "" {
		// Generate a unique human-friendly code
		raw := strings.ReplaceAll(uuid.New().String(), "-", "")
		if len(raw) > 10 {
			raw = raw[:10]
		}
		b.Code = "BK-" + strings.ToUpper(raw)
	}
	return nil
}

type BookingItem struct {
	ID            string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BookingID     string `gorm:"index" json:"booking_id"`
	RoomTypeID    int    `gorm:"index" json:"room_type_id"`
	Quantity      int    `json:"quantity"`
	PricePerNight int64  `json:"price_per_night"`
	LineTotal     int64  `json:"line_total"`
}

func (bi *BookingItem) BeforeCreate(_ *gorm.DB) error {
	if bi.ID == "" {
		bi.ID = uuid.New().String()
	}
	return nil
}

type CreateBookingItem struct {
	RoomTypeID int `json:"room_type_id" binding:"required"`
	Quantity   int `json:"quantity" binding:"required,min=1"`
}

type CreateBookingInput struct {
	UserID   string              // set by handler from JWT
	CheckIn  time.Time           `json:"check_in" binding:"required"`
	CheckOut time.Time           `json:"check_out" binding:"required"`
	Guests   int                 `json:"guests"`
	FullName string              `json:"full_name"`
	Email    string              // set by handler from JWT
	Items    []CreateBookingItem `json:"items" binding:"required,min=1,dive"`
}
