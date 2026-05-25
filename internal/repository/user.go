package repository

import (
	"backend/internal/dto"
	"backend/internal/models"
	"math"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uuid.UUID) (*models.User, error)
	FindByVerificationToken(token string) (*models.User, error)
	FindByResetToken(token string) (*models.User, error)
	List(filter dto.UserFilter) ([]models.User, int64, error)
	Update(id uuid.UUID, data map[string]interface{}) error
	SetActive(id uuid.UUID, active bool) error
	Delete(id uuid.UUID) error
	AssignRole(userID uuid.UUID, role models.Role) error
	RemoveRole(userID uuid.UUID, role models.Role) error
	ReplaceRole(userID uuid.UUID, role models.Role) error
	HasRole(userID uuid.UUID, role models.Role) (bool, error)
}

type gormUserRepo struct {
	db *gorm.DB
}

func NewGormUserRepo(db *gorm.DB) UserRepository {
	return &gormUserRepo{db: db}
}

func (r *gormUserRepo) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *gormUserRepo) FindByEmail(email string) (*models.User, error) {
	var u models.User
	err := r.db.Preload("Roles").Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *gormUserRepo) FindByID(id uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.db.Preload("Roles").First(&u, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// List returns paginated users with optional filters.
func (r *gormUserRepo) List(filter dto.UserFilter) ([]models.User, int64, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := r.db.Model(&models.User{}).Preload("Roles")

	// Search by name or email
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("users.name ILIKE ? OR users.email ILIKE ?", like, like)
	}

	// Filter by active status
	if filter.IsActive != nil {
		query = query.Where("users.is_active = ?", *filter.IsActive)
	}

	// Filter by role (join role_models)
	if filter.Role != "" {
		query = query.
			Joins("JOIN user_roles ON user_roles.user_id = users.id").
			Joins("JOIN role_models ON role_models.id = user_roles.role_model_id").
			Where("role_models.name = ?", filter.Role)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []models.User
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Order("users.created_at DESC").
		Find(&users).Error
	return users, total, err
}

func (r *gormUserRepo) Update(id uuid.UUID, data map[string]interface{}) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Updates(data).Error
}

func (r *gormUserRepo) SetActive(id uuid.UUID, active bool) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("is_active", active).Error
}

func (r *gormUserRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

// AssignRole adds a role to the user (appends, does not replace).
func (r *gormUserRepo) AssignRole(userID uuid.UUID, role models.Role) error {
	var roleModel models.RoleModel
	if err := r.db.Where("name = ?", role).First(&roleModel).Error; err != nil {
		return err
	}
	return r.db.Model(&models.User{BaseModel: models.BaseModel{ID: userID}}).
		Association("Roles").Append(&roleModel)
}

// RemoveRole removes a single role from the user.
func (r *gormUserRepo) RemoveRole(userID uuid.UUID, role models.Role) error {
	var roleModel models.RoleModel
	if err := r.db.Where("name = ?", role).First(&roleModel).Error; err != nil {
		return err
	}
	return r.db.Model(&models.User{BaseModel: models.BaseModel{ID: userID}}).
		Association("Roles").Delete(&roleModel)
}

// ReplaceRole clears all roles and assigns the given one (single-role workflow).
func (r *gormUserRepo) ReplaceRole(userID uuid.UUID, role models.Role) error {
	var roleModel models.RoleModel
	if err := r.db.Where("name = ?", role).First(&roleModel).Error; err != nil {
		return err
	}
	user := &models.User{BaseModel: models.BaseModel{ID: userID}}
	if err := r.db.Model(user).Association("Roles").Replace(&roleModel); err != nil {
		return err
	}
	return nil
}

// HasRole checks if a user has the given role.
func (r *gormUserRepo) HasRole(userID uuid.UUID, role models.Role) (bool, error) {
	var count int64
	err := r.db.Table("user_roles").
		Joins("JOIN role_models ON role_models.id = user_roles.role_model_id").
		Where("user_roles.user_id = ? AND role_models.name = ?", userID, role).
		Count(&count).Error
	return count > 0, err
}

func (r *gormUserRepo) FindByVerificationToken(token string) (*models.User, error) {
	var u models.User
	err := r.db.Preload("Roles").Where("verification_token = ?", token).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *gormUserRepo) FindByResetToken(token string) (*models.User, error) {
	var u models.User
	err := r.db.Preload("Roles").Where("reset_token = ?", token).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// PagesCount helper (used in service layer).
func PagesCount(total int64, pageSize int) int {
	if pageSize == 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(pageSize)))
}
