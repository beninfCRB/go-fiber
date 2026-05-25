package repository

import (
	"backend/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	FindByToken(tokenHash string) (*models.RefreshToken, error)
	RevokeByToken(tokenHash string) error
	RevokeAllByUser(userID uuid.UUID) error
	DeleteExpired() error
}

type gormRefreshTokenRepo struct {
	db *gorm.DB
}

func NewGormRefreshTokenRepo(db *gorm.DB) RefreshTokenRepository {
	return &gormRefreshTokenRepo{db: db}
}

func (r *gormRefreshTokenRepo) Create(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *gormRefreshTokenRepo) FindByToken(tokenHash string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.Where("token = ? AND revoked = ? AND expires_at > ?", tokenHash, false, time.Now()).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *gormRefreshTokenRepo) RevokeByToken(tokenHash string) error {
	return r.db.Model(&models.RefreshToken{}).
		Where("token = ?", tokenHash).
		Update("revoked", true).Error
}

func (r *gormRefreshTokenRepo) RevokeAllByUser(userID uuid.UUID) error {
	return r.db.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true).Error
}

func (r *gormRefreshTokenRepo) DeleteExpired() error {
	return r.db.Where("expires_at < ? OR revoked = ?", time.Now(), true).
		Delete(&models.RefreshToken{}).Error
}
