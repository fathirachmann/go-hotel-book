package entity

import "time"

type Status string

const (
	StatusUnpaid    Status = "UNPAID"
	StatusPaid      Status = "PAID"
	StatusCancelled Status = "CANCELLED"
	StatusCheckedIn Status = "CHECKED_IN"
)

type Booking struct {
	ID           string
	UserID       string
	Code         string
	CheckInDate  time.Time
	CheckOutDate time.Time
	Nights       int
	Guests       int
	Subtotal     int64
	Taxes        int64
	Total        int64
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type BookingItem struct {
	RoomTypeID    string
	Quantity      int
	PricePerNight int64
	LineTotal     int64
}

type CreateBookingInput struct {
	UserID       string
	CheckInDate  time.Time
	CheckOutDate time.Time
	Items        []BookingItem
	Email        string
}
