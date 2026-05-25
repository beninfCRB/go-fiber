package repository

import (
	"backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MenuRepository interface {
	Create(menu *models.Menu) error
	FindByID(id uuid.UUID) (*models.Menu, error)
	FindAll() ([]models.Menu, error)
	Update(id uuid.UUID, data map[string]interface{}) error
	Delete(id uuid.UUID) error

	// Role-menu assignments
	GetRolesForMenu(menuID uuid.UUID) ([]models.RoleModel, error)
	AssignRoles(menuID uuid.UUID, roles []models.RoleModel) error

	// Fetch menus accessible to specific role names (for menu tree API)
	FindByRoleNames(roleNames []string) ([]models.Menu, error)
}

type gormMenuRepo struct {
	db *gorm.DB
}

func NewGormMenuRepo(db *gorm.DB) MenuRepository {
	return &gormMenuRepo{db: db}
}

func (r *gormMenuRepo) Create(menu *models.Menu) error {
	return r.db.Create(menu).Error
}

func (r *gormMenuRepo) FindByID(id uuid.UUID) (*models.Menu, error) {
	var m models.Menu
	err := r.db.Preload("Roles").Preload("Children").First(&m, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// FindAll returns all menus (flat list) with roles and direct children preloaded.
func (r *gormMenuRepo) FindAll() ([]models.Menu, error) {
	var menus []models.Menu
	err := r.db.Preload("Roles").Preload("Children").
		Order("sort_order ASC, id ASC").
		Find(&menus).Error
	return menus, err
}

func (r *gormMenuRepo) Update(id uuid.UUID, data map[string]interface{}) error {
	return r.db.Model(&models.Menu{}).Where("id = ?", id).Updates(data).Error
}

func (r *gormMenuRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Menu{}, "id = ?", id).Error
}

func (r *gormMenuRepo) GetRolesForMenu(menuID uuid.UUID) ([]models.RoleModel, error) {
	var roles []models.RoleModel
	err := r.db.Model(&models.Menu{BaseModel: models.BaseModel{ID: menuID}}).
		Association("Roles").Find(&roles)
	return roles, err
}

// AssignRoles replaces the full set of roles for a menu.
func (r *gormMenuRepo) AssignRoles(menuID uuid.UUID, roles []models.RoleModel) error {
	menu := &models.Menu{BaseModel: models.BaseModel{ID: menuID}}
	return r.db.Model(menu).Association("Roles").Replace(roles)
}

// FindByRoleNames returns all active menus accessible to the given role names.
func (r *gormMenuRepo) FindByRoleNames(roleNames []string) ([]models.Menu, error) {
	if len(roleNames) == 0 {
		return nil, nil
	}
	var menus []models.Menu
	err := r.db.
		Joins("JOIN role_menus ON role_menus.menu_id = menus.id").
		Joins("JOIN role_models ON role_models.id = role_menus.role_model_id").
		Where("role_models.name IN ? AND menus.is_active = ?", roleNames, true).
		Where("menus.deleted_at IS NULL").
		Order("menus.sort_order ASC, menus.id ASC").
		Distinct("menus.*").
		Find(&menus).Error
	return menus, err
}
