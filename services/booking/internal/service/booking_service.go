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
	nights := daysBetween(in.CheckIn, in.CheckOut)
	if nights <= 0 {
		return nil, errors.New("invalid stay range")
	}

	if len(in.Items) == 0 {
		return nil, errors.New("booking items cannot be empty")
	}

	var subtotal int64
	var items []entity.BookingItem

	for _, it := range in.Items {
		// Simplified: snapshot first-night price
		perNight, err := s.inv.Price(it.RoomTypeID, in.CheckIn)
		if err != nil {
			return nil, err
		}

		// optional hold (currently NO-OP implementation)
		if err := s.inv.Hold(it.RoomTypeID, in.CheckIn, in.CheckOut, it.Quantity); err != nil {
			return nil, err
		}

		lineTotal := int64(it.Quantity) * int64(nights) * perNight
		subtotal += lineTotal
		items = append(items, entity.BookingItem{
			RoomTypeID:    it.RoomTypeID,
			Quantity:      it.Quantity,
			PricePerNight: perNight,
			LineTotal:     lineTotal,
		})
	}

	taxes := int64(0)
	total := subtotal + taxes

	b := &entity.Booking{
		UserID:       in.UserID,
		CheckInDate:  in.CheckIn,
		CheckOutDate: in.CheckOut,
		Nights:       nights,
		Guests:       in.Guests,
		Subtotal:     subtotal,
		Taxes:        taxes,
		Total:        total,
		Status:       entity.StatusUnpaid,
		Items:        items,
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

// RepoUpdateStatus is an internal helper to directly set booking status via repository.
func (s *Service) RepoUpdateStatus(ctx context.Context, bookingID string, status entity.Status) error {
	return s.repo.UpdateStatus(ctx, bookingID, status)
}

// ListMine returns bookings owned by the given user.
func (s *Service) ListMine(ctx context.Context, userID string) ([]entity.Booking, error) {
	return s.repo.ListByUser(ctx, userID)
}

// GetMineByID fetches a booking by id and ensures it belongs to the given user.
func (s *Service) GetMineByID(ctx context.Context, bookingID, userID string) (*entity.Booking, error) {
	b, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if b.UserID != userID {
		return nil, errors.New("forbidden")
	}
	return b, nil
}

// DeleteMine deletes a booking if it belongs to the given user.
func (s *Service) DeleteMine(ctx context.Context, bookingID, userID string) error {
	b, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}
	if b.UserID != userID {
		return errors.New("forbidden")
	}
	return s.repo.Delete(ctx, bookingID)
}
