package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	FullName       string    `gorm:"size:150;not null" json:"full_name"`
	Email          string    `gorm:"size:150;uniqueIndex;not null" json:"email"`
	HashedPassword string    `gorm:"size:255;not null" json:"-"`
	Role           string    `gorm:"size:50;not null" json:"role"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
