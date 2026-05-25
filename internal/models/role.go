package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role adalah tipe untuk nama role dalam sistem.
type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleUser       Role = "user"
)

// RoleModel adalah tabel master role.
type RoleModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        Role      `gorm:"uniqueIndex;not null;size:50"`
	Description string    `gorm:"size:255"`
	Users       []User    `gorm:"many2many:user_roles;"`
	Menus       []Menu    `gorm:"many2many:role_menus;"`
}

func (r *RoleModel) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}
