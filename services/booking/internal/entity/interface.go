package entity

import (
	"context"
	"time"
)

type InventoryRepo interface {
	Hold(roomTypeID string, checkIn, checkOut time.Time, quantity int) error
	Release(roomTypeID string, checkIn, checkOut time.Time, quantity int) error
	Price(roomTypeID string, d time.Time) (int64, error)
}

type BookingRepo interface {
	Create(ctx context.Context, b *Booking) error
	UpdateStatus(ctx context.Context, bookingID string, status Status) error
	GetByID(ctx context.Context, bookingID string) (*Booking, error)
}

type PaymentGateway interface {
	RequestPayment(ctx context.Context, bookingID string, amount int64, userEmail string) error
	RefundPayment(ctx context.Context, bookingID string, amount int64, reason string) error
}
