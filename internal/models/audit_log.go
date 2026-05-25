package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog merepresentasikan pencatatan aktivitas pengguna untuk audit log keamanan.
type AuditLog struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Action      string     `gorm:"type:varchar(100);not null" json:"action"`
	Description string     `gorm:"type:text;not null" json:"description"`
	IPAddress   string     `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent   string     `gorm:"type:text" json:"user_agent"`
	CreatedAt   time.Time  `json:"created_at"`
}
