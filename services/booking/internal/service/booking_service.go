package service

import (
	"booking/internal/entity"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	inv  entity.InventoryRepo
	repo entity.BookingRepo
	pay  entity.PaymentGateway
}

var (
	// ErrBookingNotFound signals caller the booking cannot be located.
	ErrBookingNotFound = errors.New("booking not found")
	// ErrInvalidBookingStatus indicates the booking is not in the expected state for the requested mutation.
	ErrInvalidBookingStatus = errors.New("booking status does not allow this action")
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
	var total int64
	for d := 0; d < nights; d++ {
		day := in.CheckInDate.AddDate(0, 0, d)
		for _, it := range in.Items {
			if err := s.inv.Hold(it.RoomTypeID, in.CheckInDate, in.CheckOutDate, it.Quantity); err != nil {
				return nil, err
			}
			p, err := s.inv.Price(it.RoomTypeID, day)
			if err != nil {
				return nil, err
			}
			total += p * int64(it.Quantity)
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}

	if booking.Status != entity.StatusPaid {
		return nil, ErrInvalidBookingStatus
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}

	if booking.Status != entity.StatusPaid {
		return nil, ErrInvalidBookingStatus
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
