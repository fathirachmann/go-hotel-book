package service

import (
	"catalog/internal/entity"
	"catalog/internal/repo"
	"context"
	"fmt"
	"time"
)

// AvailabilityItem represents the availability response for a room type.
type AvailabilityItem struct {
	RoomTypeID    string `json:"room_type_id"`
	Name          string `json:"name"`
	Capacity      int    `json:"capacity"`
	Available     int    `json:"available"`
	PricePerNight int64  `json:"price_per_night"`
	TotalPrice    int64  `json:"total_price"`
}

// CatalogService orchestrates catalog business use-cases.
type CatalogService struct {
	roomTypes repo.RoomTypeRepository
	inventory repo.InventoryRepository
	clock     func() time.Time
}

// NewCatalogService wires dependencies for catalog use-cases.
func NewCatalogService(rt repo.RoomTypeRepository, inv repo.InventoryRepository) *CatalogService {
	return &CatalogService{
		roomTypes: rt,
		inventory: inv,
		clock:     time.Now,
	}
}

func daysBetween(from, to time.Time) int {
	n := int(to.Sub(from).Hours() / 24)
	if n < 0 {
		return 0
	}
	return n
}

// SeedSample seeds basic room types and inventory window for quick demos.
func (s *CatalogService) SeedSample(ctx context.Context) error {
	// Reset existing seed data (inventories first, then room types)
	if err := s.inventory.DeleteAll(ctx); err != nil {
		return err
	}
	if err := s.roomTypes.DeleteAll(ctx); err != nil {
		return err
	}

	// Seed some basic room types
	samples := []entity.RoomType{
		{Name: "Deluxe", Description: "Queen bed", BasePrice: 750000, Capacity: 2},
		{Name: "Suite", Description: "King bed + living area", BasePrice: 1550000, Capacity: 3},
		{Name: "Family", Description: "2 Queen beds", BasePrice: 1200000, Capacity: 4},
		{Name: "Standard", Description: "Cozy room", BasePrice: 550000, Capacity: 2},
	}

	for i := range samples {
		if err := s.roomTypes.Upsert(ctx, &samples[i]); err != nil {
			return err
		}
	}

	types, err := s.roomTypes.List(ctx)
	if err != nil {
		return err
	}

	// Seed next 30 days of inventory with simple weekend price overrides (+15%)
	today := s.clock().Truncate(24 * time.Hour)
	nights := 30
	for _, rt := range types {
		for i := 0; i < nights; i++ {
			day := today.AddDate(0, 0, i)
			inv := entity.RoomInventory{
				RoomTypeID:     rt.ID,
				InvDate:        day,
				TotalRooms:     10,
				AvailableRooms: 10,
			}

			// Weekend override: Friday or Saturday +15%
			wd := day.Weekday()
			if wd == time.Friday || wd == time.Saturday {
				override := int64(float64(rt.BasePrice) * 1.15)
				inv.PriceOverride = &override
			}

			if err := s.inventory.Upsert(ctx, &inv); err != nil {
				return err
			}
		}
	}

	return nil
}

// Availability returns room types available for the supplied range and guest count.
func (s *CatalogService) Availability(ctx context.Context, from, to time.Time, guests int) ([]AvailabilityItem, error) {
	nights := daysBetween(from, to)
	if nights <= 0 {
		return []AvailabilityItem{}, nil
	}

	types, err := s.roomTypes.List(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]AvailabilityItem, 0, len(types))
	for _, rt := range types {
		if guests > 0 && rt.Capacity < guests {
			continue
		}

		minAvail, err := s.inventory.MinAvailable(ctx, rt.ID, from, to)
		if err != nil {
			return nil, err
		}
		if minAvail <= 0 {
			continue
		}

		overrides, err := s.inventory.Prices(ctx, rt.ID, from, to)
		if err != nil {
			return nil, err
		}

		var (
			pricePerNight = rt.BasePrice
			total         int64
		)

		for i := 0; i < nights; i++ {
			price := rt.BasePrice
			if i < len(overrides) && overrides[i] >= 0 {
				price = overrides[i]
			}
			if i == 0 {
				pricePerNight = price
			}
			total += price
		}

		items = append(items, AvailabilityItem{
			RoomTypeID:    fmt.Sprint(rt.ID),
			Name:          rt.Name,
			Capacity:      rt.Capacity,
			Available:     minAvail,
			PricePerNight: pricePerNight,
			TotalPrice:    total,
		})
	}

	return items, nil
}
