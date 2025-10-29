package entity

import (
	"context"
	"time"
)

type InventoryRepo interface {
	Hold(roomTypeID int, checkIn, checkOut time.Time, quantity int) error
	Release(roomTypeID int, checkIn, checkOut time.Time, quantity int) error
	Price(roomTypeID int, d time.Time) (int64, error)
}

type BookingRepo interface {
	Create(ctx context.Context, b *Booking) error
	UpdateStatus(ctx context.Context, bookingID string, status Status) error
	GetByID(ctx context.Context, bookingID string) (*Booking, error)
	ListByUser(ctx context.Context, userID string) ([]Booking, error)
}

type PaymentGateway interface {
	RequestPayment(ctx context.Context, bookingID string, amount int64, userEmail string) error
	RefundPayment(ctx context.Context, bookingID string, amount int64, reason string) error
}
