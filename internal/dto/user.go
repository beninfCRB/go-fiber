package dto

import (
	"backend/internal/models"

	"github.com/google/uuid"
)

// ── Request DTOs ─────────────────────────────────────────────────────────────

type CreateUserRequest struct {
	Name     string       `json:"name"     validate:"required,min=2"`
	Email    string       `json:"email"    validate:"required,email"`
	Password string       `json:"password" validate:"required,min=8"`
	Role     models.Role  `json:"role"     validate:"required,oneof=super_admin admin user"`
}

type UpdateUserRequest struct {
	Name     string `json:"name"  validate:"omitempty,min=2"`
	Email    string `json:"email" validate:"omitempty,email"`
	Password string `json:"password" validate:"omitempty,min=8"`
}

type AssignRoleRequest struct {
	Role models.Role `json:"role" validate:"required,oneof=super_admin admin user"`
}

type ToggleActiveRequest struct {
	IsActive bool `json:"is_active"`
}

// ── Query / Filter DTOs ───────────────────────────────────────────────────────

type UserFilter struct {
	Role     string `query:"role"`
	IsActive *bool  `query:"is_active"`
	Search   string `query:"search"` // search by name or email
	Page     int    `query:"page"    validate:"omitempty,min=1"`
	PageSize int    `query:"page_size" validate:"omitempty,min=1,max=100"`
}

// ── Response DTOs ─────────────────────────────────────────────────────────────

type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	IsActive bool      `json:"is_active"`
	Roles    []string  `json:"roles"`
}

type UserListResponse struct {
	Data       []UserResponse `json:"data"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}
