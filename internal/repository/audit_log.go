package repository

import (
	"backend/internal/models"

	"gorm.io/gorm"
)

// AuditLogRepository menyediakan operasi database untuk merekam aktivitas audit pengguna.
type AuditLogRepository interface {
	Create(log *models.AuditLog) error
	List(page, pageSize int) ([]models.AuditLog, int64, error)
}

type gormAuditLogRepo struct {
	db *gorm.DB
}

// NewGormAuditLogRepo membuat instance baru dari repository AuditLog.
func NewGormAuditLogRepo(db *gorm.DB) AuditLogRepository {
	return &gormAuditLogRepo{db: db}
}

func (r *gormAuditLogRepo) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *gormAuditLogRepo) List(page, pageSize int) ([]models.AuditLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	if err := r.db.Model(&models.AuditLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []models.AuditLog
	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}
