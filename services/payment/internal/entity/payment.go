package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentStatus string

const (
	PayPending    PaymentStatus = "PENDING"
	PaySettlement PaymentStatus = "SETTLEMENT"
	PayExpire     PaymentStatus = "EXPIRE"
	PayDeny       PaymentStatus = "DENY"
	PayRefunded   PaymentStatus = "REFUNDED"
)

type Payment struct {
	ID          string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BookingID   string `gorm:"index"`
	OrderID     string `gorm:"uniqueIndex"`
	Amount      int64
	Provider    string
	ProviderRef string
	Status      PaymentStatus `gorm:"index"`
	RawPayload  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Refund struct {
	ID        string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PaymentID string `gorm:"index"`
	Amount    int64
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Hooks to ensure UUIDs are present even if DB default is unavailable
func (p *Payment) BeforeCreate(_ *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	return nil
}

func (r *Refund) BeforeCreate(_ *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
	return nil
}
