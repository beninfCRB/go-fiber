package service

import (
	"backend/internal/models"
	"backend/internal/repository"
	"time"

	"github.com/google/uuid"
)

type AuditLogService struct {
	repo repository.AuditLogRepository
}

func NewAuditLogService(repo repository.AuditLogRepository) *AuditLogService {
	return &AuditLogService{repo: repo}
}

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
