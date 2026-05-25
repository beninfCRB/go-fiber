package models

import (
	"time"
)

type User struct {
	BaseModel
	Name                 string      `gorm:"not null"`
	Email                string      `gorm:"uniqueIndex;not null"`
	Password             string      `gorm:"not null"`
	IsActive             bool        `gorm:"default:true"`
	IsVerified           bool        `gorm:"default:false"`
	VerificationToken    string      `gorm:"size:255;index"`
	ResetToken           string      `gorm:"size:255;index"`
	ResetTokenExpiresAt  *time.Time
	Roles                []RoleModel `gorm:"many2many:user_roles;"`
}
