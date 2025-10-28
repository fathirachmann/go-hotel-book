package service

import (
	"booking/internal/entity"
	"context"
	"errors"
	"time"
)

type Service struct {
	inv  entity.InventoryRepo
	repo entity.BookingRepo
	pay  entity.PaymentGateway
}

var (
	// ErrBookingNotPaid is returned when booking is not in paid state.
	ErrBookingNotPaid = errors.New("booking is not paid yet")
	// ErrBookingAlreadyHandled is returned when booking already cancelled or checked-in.
	ErrBookingAlreadyHandled = errors.New("booking already handled")
)

func NewService(inv entity.InventoryRepo, repo entity.BookingRepo, pay entity.PaymentGateway) *Service {
	return &Service{
		inv:  inv,
		repo: repo,
		pay:  pay,
	}
}

func daysBetween(ci, co time.Time) int {
	d := int(co.Sub(ci).Hours() / 24)
	if d < 0 {
		return 0
	}
	return d
}

func (s *Service) Create(ctx context.Context, in entity.CreateBookingInput) (*entity.Booking, error) {
	nights := daysBetween(in.CheckInDate, in.CheckOutDate)
	if nights <= 0 {
		return nil, errors.New("invalid stay range")
	}

	if len(in.Items) == 0 {
		return nil, errors.New("booking items cannot be empty")
	}

	var total int64

	for _, it := range in.Items {
		// reserve stock for the whole stay
		if err := s.inv.Hold(it.RoomTypeID, in.CheckInDate, in.CheckOutDate, it.Quantity); err != nil {
			return nil, err
		}

		for d := 0; d < nights; d++ {
			day := in.CheckInDate.AddDate(0, 0, d)
			price, err := s.inv.Price(it.RoomTypeID, day)
			if err != nil {
				return nil, err
			}
			total += price * int64(it.Quantity)
		}
	}

	b := &entity.Booking{
		UserID:       in.UserID,
		CheckInDate:  in.CheckInDate,
		CheckOutDate: in.CheckOutDate,
		Nights:       nights,
		Total:        total,
		Status:       entity.StatusUnpaid,
	}

	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}

	_ = s.pay.RequestPayment(ctx, b.ID, b.Total, in.Email)

	return b, nil
}

// CheckIn marks a booking as checked-in when payment is settled.
func (s *Service) CheckIn(ctx context.Context, bookingID string) (*entity.Booking, error) {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if booking.Status == entity.StatusCheckedIn {
		return nil, ErrBookingAlreadyHandled
	}

	if booking.Status != entity.StatusPaid {
		return nil, ErrBookingNotPaid
	}

	if err := s.repo.UpdateStatus(ctx, booking.ID, entity.StatusCheckedIn); err != nil {
		return nil, err
	}
	booking.Status = entity.StatusCheckedIn
	return booking, nil
}

// Refund triggers refund flow via payment service and cancels the booking.
func (s *Service) Refund(ctx context.Context, bookingID, reason string) (*entity.Booking, error) {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if booking.Status == entity.StatusCancelled {
		return nil, ErrBookingAlreadyHandled
	}

	if booking.Status != entity.StatusPaid {
		return nil, ErrBookingNotPaid
	}

	if reason == "" {
		reason = "user requested"
	}

	if err := s.pay.RefundPayment(ctx, booking.ID, booking.Total, reason); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateStatus(ctx, booking.ID, entity.StatusCancelled); err != nil {
		return nil, err
	}
	booking.Status = entity.StatusCancelled
	return booking, nil
}
