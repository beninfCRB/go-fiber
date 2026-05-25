package service

import (
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/repository"
	"errors"
	"math"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound          = errors.New("user not found")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrCannotManageSelf  = errors.New("cannot modify your own account through this endpoint")
	ErrRoleNotAllowed    = errors.New("you are not allowed to assign this role")
	ErrEmailTaken        = errors.New("email already taken")
)

// UserService handles CRUD operations with RBAC enforcement.
// Rules:
//   - super_admin → can manage admin + user (NOT other super_admins unless self)
//   - admin       → can only manage users (role=user), cannot touch admins
type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// primaryRoleOf returns the highest role of a user model.
func primaryRoleOf(u *models.User) models.Role {
	priority := map[models.Role]int{
		models.RoleSuperAdmin: 3,
		models.RoleAdmin:      2,
		models.RoleUser:       1,
	}
	best := models.RoleUser
	bestScore := 0
	for _, r := range u.Roles {
		if s := priority[r.Name]; s > bestScore {
			bestScore = s
			best = r.Name
		}
	}
	return best
}

// canManage checks whether the requester can act on the target.
func canManage(requesterRole models.Role, targetRole models.Role) bool {
	switch requesterRole {
	case models.RoleSuperAdmin:
		// super_admin can manage admin and user, not another super_admin
		return targetRole != models.RoleSuperAdmin
	case models.RoleAdmin:
		// admin can only manage users
		return targetRole == models.RoleUser
	}
	return false
}

// canAssignRole checks whether the requester is allowed to assign a specific role.
func canAssignRole(requesterRole models.Role, target models.Role) bool {
	switch requesterRole {
	case models.RoleSuperAdmin:
		return target == models.RoleAdmin || target == models.RoleUser
	case models.RoleAdmin:
		return target == models.RoleUser
	}
	return false
}

// toUserResponse converts a models.User to the public DTO.
func toUserResponse(u *models.User) dto.UserResponse {
	roles := make([]string, 0, len(u.Roles))
	for _, r := range u.Roles {
		roles = append(roles, string(r.Name))
	}
	return dto.UserResponse{
		ID:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		IsActive: u.IsActive,
		Roles:    roles,
	}
}

// ── Service methods ───────────────────────────────────────────────────────────

// ListUsers returns paginated users visible to the requester.
// admin  → restricted to role=user only.
// super_admin → can filter any role.
func (s *UserService) ListUsers(requesterRole models.Role, filter dto.UserFilter) (*dto.UserListResponse, error) {
	// Admins may only see regular users
	if requesterRole == models.RoleAdmin {
		filter.Role = string(models.RoleUser)
	}

	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	page := filter.Page
	if page < 1 {
		page = 1
	}

	users, total, err := s.userRepo.List(filter)
	if err != nil {
		return nil, err
	}

	data := make([]dto.UserResponse, 0, len(users))
	for _, u := range users {
		data = append(data, toUserResponse(&u))
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &dto.UserListResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetUser returns a single user if the requester is allowed to view them.
func (s *UserService) GetUser(requesterRole models.Role, requesterID, targetID uuid.UUID) (*dto.UserResponse, error) {
	target, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return nil, ErrNotFound
	}
	targetRole := primaryRoleOf(target)
	if requesterID != targetID && !canManage(requesterRole, targetRole) {
		return nil, ErrPermissionDenied
	}
	resp := toUserResponse(target)
	return &resp, nil
}

// CreateUser creates a new user and assigns the given role.
func (s *UserService) CreateUser(requesterRole models.Role, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	if !canAssignRole(requesterRole, req.Role) {
		return nil, ErrRoleNotAllowed
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hash),
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, ErrEmailTaken
	}
	if err := s.userRepo.AssignRole(user.ID, req.Role); err != nil {
		return nil, err
	}

	created, _ := s.userRepo.FindByID(user.ID)
	resp := toUserResponse(created)
	return &resp, nil
}

// UpdateUser updates name/email/password of a target user.
func (s *UserService) UpdateUser(requesterRole models.Role, requesterID, targetID uuid.UUID, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	target, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return nil, ErrNotFound
	}
	targetRole := primaryRoleOf(target)
	if requesterID != targetID && !canManage(requesterRole, targetRole) {
		return nil, ErrPermissionDenied
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Password != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		updates["password"] = string(hash)
	}

	if len(updates) > 0 {
		if err := s.userRepo.Update(targetID, updates); err != nil {
			return nil, err
		}
	}

	updated, _ := s.userRepo.FindByID(targetID)
	resp := toUserResponse(updated)
	return &resp, nil
}

// AssignRole replaces all roles of the target user with a single new role.
func (s *UserService) AssignRole(requesterRole models.Role, requesterID, targetID uuid.UUID, newRole models.Role) (*dto.UserResponse, error) {
	if requesterID == targetID {
		return nil, ErrCannotManageSelf
	}
	target, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return nil, ErrNotFound
	}
	if !canManage(requesterRole, primaryRoleOf(target)) {
		return nil, ErrPermissionDenied
	}
	if !canAssignRole(requesterRole, newRole) {
		return nil, ErrRoleNotAllowed
	}

	if err := s.userRepo.ReplaceRole(targetID, newRole); err != nil {
		return nil, err
	}

	updated, _ := s.userRepo.FindByID(targetID)
	resp := toUserResponse(updated)
	return &resp, nil
}

// SetActive activates or deactivates a user.
func (s *UserService) SetActive(requesterRole models.Role, requesterID, targetID uuid.UUID, active bool) error {
	if requesterID == targetID {
		return ErrCannotManageSelf
	}
	target, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return ErrNotFound
	}
	if !canManage(requesterRole, primaryRoleOf(target)) {
		return ErrPermissionDenied
	}
	return s.userRepo.SetActive(targetID, active)
}

// DeleteUser soft-deletes a user.
func (s *UserService) DeleteUser(requesterRole models.Role, requesterID, targetID uuid.UUID) error {
	if requesterID == targetID {
		return ErrCannotManageSelf
	}
	target, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return ErrNotFound
	}
	if !canManage(requesterRole, primaryRoleOf(target)) {
		return ErrPermissionDenied
	}
	return s.userRepo.Delete(targetID)
}
