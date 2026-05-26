package dto

import (
	"github.com/google/uuid"
)

// ── Menu Request DTOs ──────────────────────────────────────────────────────────

type CreateMenuRequest struct {
	Name      string     `json:"name"       validate:"required,min=2"`
	Key       string     `json:"key"        validate:"required,min=2"`
	Path      string     `json:"path"       validate:"required"`
	Icon      string     `json:"icon"`
	ParentID  *uuid.UUID `json:"parent_id"`
	SortOrder int        `json:"sort_order"`
	IsActive  bool       `json:"is_active"`
	RoleKeys  []string   `json:"role_keys"` // e.g. ["admin","user"]
}

type UpdateMenuRequest struct {
	Name      string     `json:"name"       validate:"omitempty,min=2"`
	Path      string     `json:"path"`
	Icon      string     `json:"icon"`
	ParentID  *uuid.UUID `json:"parent_id"`
	SortOrder *int       `json:"sort_order"`
	IsActive  *bool      `json:"is_active"`
}

type AssignMenuRolesRequest struct {
	RoleKeys []string `json:"role_keys" validate:"required,min=1"`
}

// ── Menu Response DTOs ─────────────────────────────────────────────────────────

type MenuResponse struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	Key       string         `json:"key"`
	Path      string         `json:"path"`
	Icon      string         `json:"icon"`
	ParentID  *uuid.UUID     `json:"parent_id,omitempty"`
	SortOrder int            `json:"sort_order"`
	IsActive  bool           `json:"is_active"`
	Roles     []string       `json:"roles,omitempty"`    // visible in admin/super-admin view
	Children  []MenuResponse `json:"children,omitempty"` // nested menus
}

type MenuTreeResponse struct {
	Menus []MenuResponse `json:"menus"`
}
