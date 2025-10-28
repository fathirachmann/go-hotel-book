package service

import (
	"booking/internal/entity"
	"context"
	"time"
)

type Service struct {
	inv  entity.InventoryRepo
	repo entity.BookingRepo
	pay  entity.PaymentGateway
}

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
