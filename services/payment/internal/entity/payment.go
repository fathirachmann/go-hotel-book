package entity

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Status represents the current payment state within our system.
type Status string

const (
	StatusPending   Status = "PENDING"
	StatusPaid      Status = "PAID"
	StatusFailed    Status = "FAILED"
	StatusCancelled Status = "CANCELLED"
	StatusExpired   Status = "EXPIRED"
)

// ParseStatus normalises and validates a status string.
func ParseStatus(v string) (Status, error) {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case string(StatusPending):
		return StatusPending, nil
	case string(StatusPaid):
		return StatusPaid, nil
	case string(StatusFailed):
		return StatusFailed, nil
	case string(StatusCancelled):
		return StatusCancelled, nil
	case string(StatusExpired):
		return StatusExpired, nil
	default:
		return "", fmt.Errorf("unknown payment status: %s", v)
	}
}

// Payment captures outbound requests and asynchronous updates from payment provider.
type Payment struct {
	ID               uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	BookingID        string          `gorm:"size:64;index;not null" json:"booking_id"`
	OrderID          string          `gorm:"size:100;uniqueIndex;not null" json:"order_id"`
	Provider         string          `gorm:"size:30;not null" json:"provider"`
	Amount           int64           `gorm:"not null" json:"amount"`
	Currency         string          `gorm:"size:5;not null" json:"currency"`
	Status           Status          `gorm:"size:20;not null" json:"status"`
	RedirectURL      string          `gorm:"size:255" json:"redirect_url"`
	SnapToken        string          `gorm:"size:120" json:"snap_token"`
	CustomerEmail    string          `gorm:"size:150" json:"customer_email"`
	CustomerName     string          `gorm:"size:150" json:"customer_name"`
	LastNotification json.RawMessage `gorm:"type:jsonb" json:"-"`
	PaidAt           *time.Time      `json:"paid_at"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// BeforeCreate populates UUID primary key.
func (p *Payment) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
