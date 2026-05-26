package models

import (
	"github.com/google/uuid"
)

type Menu struct {
	BaseModel
	Name      string     `gorm:"not null"`
	Key       string     `gorm:"uniqueIndex;not null"` // unique slug, e.g. "user-management"
	Path      string     `gorm:"not null"`             // frontend route, e.g. "/admin/users"
	Icon      string     `gorm:"default:''"`           // icon name (e.g. "users", "shield")
	ParentID  *uuid.UUID `gorm:"type:uuid;index"`      // nil = root menu
	SortOrder int        `gorm:"default:0"`
	IsActive  bool       `gorm:"default:true"`
	// Relations
	Roles    []RoleModel `gorm:"many2many:role_menus;"`
	Children []Menu      `gorm:"foreignKey:ParentID"`
}
