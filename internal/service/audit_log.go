package service

import (
	"backend/internal/models"
	"backend/internal/repository"
	"time"

	"github.com/google/uuid"
)

// AuditLogService menangani pencatatan log audit keamanan secara terpusat.
type AuditLogService struct {
	repo repository.AuditLogRepository
}

// NewAuditLogService membuat instance baru dari AuditLogService.
func NewAuditLogService(repo repository.AuditLogRepository) *AuditLogService {
	return &AuditLogService{repo: repo}
}

// LogAction mencatat aksi pengguna ke tabel audit_logs.
func (s *AuditLogService) LogAction(userID *uuid.UUID, action, description, ipAddress, userAgent string) error {
	log := &models.AuditLog{
		ID:          uuid.New(),
		UserID:      userID,
		Action:      action,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		CreatedAt:   time.Now(),
	}
	return s.repo.Create(log)
}
