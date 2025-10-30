package entity

import "context"

// PaymentRepo defines storage operations for Payment entities.
type PaymentRepo interface {
	Create(ctx context.Context, p *Payment) error
	FindByOrderID(ctx context.Context, orderID string) (*Payment, error)
	UpdateStatus(ctx context.Context, id string, status PaymentStatus, raw string, providerRef string) error
	// ListByUserID returns all payments for bookings owned by the given user ID
	ListByUserID(ctx context.Context, userID string) ([]Payment, error)
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
