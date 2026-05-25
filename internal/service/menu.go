package service

import (
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/repository"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrMenuNotFound  = errors.New("menu not found")
	ErrMenuKeyTaken  = errors.New("menu key already exists")
	ErrRoleNotFound  = errors.New("role not found in the system")
)

type MenuService struct {
	menuRepo repository.MenuRepository
	roleRepo repository.RoleRepository
}

func NewMenuService(menuRepo repository.MenuRepository, roleRepo repository.RoleRepository) *MenuService {
	return &MenuService{menuRepo: menuRepo, roleRepo: roleRepo}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func toMenuResponse(m *models.Menu, includeRoles bool) dto.MenuResponse {
	roles := []string{}
	if includeRoles {
		for _, r := range m.Roles {
			roles = append(roles, string(r.Name))
		}
	}
	children := []dto.MenuResponse{}
	for _, c := range m.Children {
		children = append(children, toMenuResponse(&c, false))
	}
	return dto.MenuResponse{
		ID:        m.ID,
		Name:      m.Name,
		Key:       m.Key,
		Path:      m.Path,
		Icon:      m.Icon,
		ParentID:  m.ParentID,
		SortOrder: m.SortOrder,
		IsActive:  m.IsActive,
		Roles:     roles,
		Children:  children,
	}
}

// buildTree converts a flat menu list into a nested tree.
func buildTree(menus []models.Menu, parentID *uuid.UUID) []dto.MenuResponse {
	result := []dto.MenuResponse{}
	for _, m := range menus {
		// Match parentID (both nil, or same value)
		if (m.ParentID == nil && parentID == nil) ||
			(m.ParentID != nil && parentID != nil && *m.ParentID == *parentID) {
			node := toMenuResponse(&m, false)
			node.Children = buildTree(menus, &m.ID)
			result = append(result, node)
		}
	}
	return result
}

// ── Service methods ───────────────────────────────────────────────────────────

// GetUserMenuTree returns the menu tree for a user based on their role names.
// This is what the frontend calls on login to build the sidebar.
func (s *MenuService) GetUserMenuTree(roleNames []string) ([]dto.MenuResponse, error) {
	menus, err := s.menuRepo.FindByRoleNames(roleNames)
	if err != nil {
		return nil, err
	}
	return buildTree(menus, nil), nil
}

// ListAllMenus returns every menu (flat, with roles), used by super_admin.
func (s *MenuService) ListAllMenus() ([]dto.MenuResponse, error) {
	menus, err := s.menuRepo.FindAll()
	if err != nil {
		return nil, err
	}
	result := make([]dto.MenuResponse, 0, len(menus))
	for _, m := range menus {
		result = append(result, toMenuResponse(&m, true))
	}
	return result, nil
}

// GetMenuTree returns the full menu tree, used by super_admin to see structure.
func (s *MenuService) GetMenuTree() ([]dto.MenuResponse, error) {
	menus, err := s.menuRepo.FindAll()
	if err != nil {
		return nil, err
	}
	return buildTree(menus, nil), nil
}

// GetMenu returns a single menu with its roles.
func (s *MenuService) GetMenu(id uuid.UUID) (*dto.MenuResponse, error) {
	m, err := s.menuRepo.FindByID(id)
	if err != nil {
		return nil, ErrMenuNotFound
	}
	resp := toMenuResponse(m, true)
	return &resp, nil
}

// CreateMenu creates a new menu item and assigns it to the given roles.
func (s *MenuService) CreateMenu(req dto.CreateMenuRequest) (*dto.MenuResponse, error) {
	menu := &models.Menu{
		Name:      req.Name,
		Key:       req.Key,
		Path:      req.Path,
		Icon:      req.Icon,
		ParentID:  req.ParentID,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	}
	if err := s.menuRepo.Create(menu); err != nil {
		return nil, ErrMenuKeyTaken
	}

	if len(req.RoleKeys) > 0 {
		roles, err := s.roleRepo.FindByNames(req.RoleKeys)
		if err != nil || len(roles) == 0 {
			return nil, ErrRoleNotFound
		}
		if err := s.menuRepo.AssignRoles(menu.ID, roles); err != nil {
			return nil, err
		}
	}

	created, _ := s.menuRepo.FindByID(menu.ID)
	resp := toMenuResponse(created, true)
	return &resp, nil
}

// UpdateMenu updates mutable fields of a menu.
func (s *MenuService) UpdateMenu(id uuid.UUID, req dto.UpdateMenuRequest) (*dto.MenuResponse, error) {
	if _, err := s.menuRepo.FindByID(id); err != nil {
		return nil, ErrMenuNotFound
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Path != "" {
		updates["path"] = req.Path
	}
	if req.Icon != "" {
		updates["icon"] = req.Icon
	}
	if req.ParentID != nil {
		updates["parent_id"] = req.ParentID
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) > 0 {
		if err := s.menuRepo.Update(id, updates); err != nil {
			return nil, err
		}
	}

	updated, _ := s.menuRepo.FindByID(id)
	resp := toMenuResponse(updated, true)
	return &resp, nil
}

// AssignMenuRoles replaces which roles can access a menu.
func (s *MenuService) AssignMenuRoles(menuID uuid.UUID, roleKeys []string) (*dto.MenuResponse, error) {
	if _, err := s.menuRepo.FindByID(menuID); err != nil {
		return nil, ErrMenuNotFound
	}

	roles, err := s.roleRepo.FindByNames(roleKeys)
	if err != nil {
		return nil, err
	}

	if err := s.menuRepo.AssignRoles(menuID, roles); err != nil {
		return nil, err
	}

	updated, _ := s.menuRepo.FindByID(menuID)
	resp := toMenuResponse(updated, true)
	return &resp, nil
}

// DeleteMenu soft-deletes a menu item.
func (s *MenuService) DeleteMenu(id uuid.UUID) error {
	if _, err := s.menuRepo.FindByID(id); err != nil {
		return ErrMenuNotFound
	}
	return s.menuRepo.Delete(id)
}
