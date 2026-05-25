package models

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	BaseModel
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Revoked   bool      `gorm:"default:false"`
	User      User      `gorm:"foreignKey:UserID"`
}
