package entity

import "context"

// PaymentRepo defines storage operations for Payment entities.
type PaymentRepo interface {
	Create(ctx context.Context, p *Payment) error
	FindByOrderID(ctx context.Context, orderID string) (*Payment, error)
	UpdateStatus(ctx context.Context, id string, status PaymentStatus, raw string, providerRef string) error
	ListByUserID(ctx context.Context, userID string) ([]Payment, error)
	// GetBookingTotal returns the total price of a booking from booking schema
	GetBookingTotal(ctx context.Context, bookingID string) (int64, error)
}

// RefundRepo defines storage operations for Refund entities.
type RefundRepo interface {
	Create(ctx context.Context, r *Refund) error
}

// BookingClient abstracts calls to the Booking service for status updates.
type BookingClient interface {
	UpdateStatusPaid(ctx context.Context, bookingID string) error
	UpdateStatusExpired(ctx context.Context, bookingID string) error
	UpdateStatusRefunded(ctx context.Context, bookingID string) error
}
