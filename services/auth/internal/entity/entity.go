package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents an application account stored in the auth schema.
type User struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	FullName       string    `gorm:"size:150;not null" json:"full_name"`
	Email          string    `gorm:"size:150;uniqueIndex;not null" json:"email"`
	HashedPassword string    `gorm:"size:255;not null" json:"-"`
	Role           string    `gorm:"size:50;not null" json:"role"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// BeforeCreate ensures a UUID is assigned when the record is first inserted.
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
