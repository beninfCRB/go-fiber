package repository

import (
	"backend/internal/models"

	"gorm.io/gorm"
)

type RoleRepository interface {
	FindAll() ([]models.RoleModel, error)
	FindByName(name string) (*models.RoleModel, error)
	FindByNames(names []string) ([]models.RoleModel, error)
}

type gormRoleRepo struct {
	db *gorm.DB
}

func NewGormRoleRepo(db *gorm.DB) RoleRepository {
	return &gormRoleRepo{db: db}
}

func (r *gormRoleRepo) FindAll() ([]models.RoleModel, error) {
	var roles []models.RoleModel
	err := r.db.Preload("Menus").Find(&roles).Error
	return roles, err
}

func (r *gormRoleRepo) FindByName(name string) (*models.RoleModel, error) {
	var role models.RoleModel
	err := r.db.Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *gormRoleRepo) FindByNames(names []string) ([]models.RoleModel, error) {
	var roles []models.RoleModel
	err := r.db.Where("name IN ?", names).Find(&roles).Error
	return roles, err
}
