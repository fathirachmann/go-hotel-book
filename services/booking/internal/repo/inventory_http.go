package repo

import (
	"booking/internal/entity"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type InventoryHTTP struct {
	base   string
	client *http.Client
}

func NewInventoryHTTPRepo(baseURL string) entity.InventoryRepo {
	return &InventoryHTTP{
		base:   baseURL,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (r *InventoryHTTP) Hold(roomTypeID int, from, to time.Time, qty int) error    { return nil }
func (r *InventoryHTTP) Release(roomTypeID int, from, to time.Time, qty int) error { return nil }

type catalogAvailabilityItem struct {
	RoomTypeID    int   `json:"room_type_id"`
	PricePerNight int64 `json:"price_per_night"`
}
type catalogAvailabilityResp struct {
	Data []catalogAvailabilityItem `json:"data"`
}

func (r *InventoryHTTP) Price(roomTypeID int, d time.Time) (int64, error) {
	// Query one-day range [d, d+1)
	q := url.Values{}
	q.Set("check_in", d.Format("2006-01-02"))
	q.Set("check_out", d.AddDate(0, 0, 1).Format("2006-01-02"))
	u := fmt.Sprintf("%s/catalog/availability?%s", r.base, q.Encode())

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, u, nil)
	resp, err := r.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("catalog returned %d", resp.StatusCode)
	}
	var out catalogAvailabilityResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, err
	}
	for _, it := range out.Data {
		if it.RoomTypeID == roomTypeID {
			return it.PricePerNight, nil
		}
	}
	return 0, fmt.Errorf("room_type_id %d not found", roomTypeID)
}
