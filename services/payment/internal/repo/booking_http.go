package repo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type bookingHTTP struct {
	base string
	cli  *http.Client
}

func NewBookingHTTPClient(base string) *bookingHTTP {
	if base == "" {
		base = "http://booking:8003"
	}
	return &bookingHTTP{
		base: base,
		cli:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (b *bookingHTTP) postStatus(ctx context.Context, bookingID, status string) error {
	url := fmt.Sprintf("%s/internal/bookings/%s/status", b.base, bookingID)
	payload := map[string]string{"status": status}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := b.cli.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("booking status update failed: %s", res.Status)
	}
	return nil
}

func (b *bookingHTTP) UpdateStatusPaid(ctx context.Context, bookingID string) error {
	return b.postStatus(ctx, bookingID, "PAID")
}
func (b *bookingHTTP) UpdateStatusExpired(ctx context.Context, bookingID string) error {
	return b.postStatus(ctx, bookingID, "CANCELLED")
}
func (b *bookingHTTP) UpdateStatusRefunded(ctx context.Context, bookingID string) error {
	return b.postStatus(ctx, bookingID, "REFUNDED")
}
