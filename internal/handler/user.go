package handler

import (
	"backend/internal/dto"
	"backend/internal/helper"
	"backend/internal/middleware"
	"backend/internal/service"
	"errors"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	userService *service.UserService
	auditLog    *service.AuditLogService
}

func NewUserHandler(svc *service.UserService, auditLog *service.AuditLogService) *UserHandler {
	return &UserHandler{userService: svc, auditLog: auditLog}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func handleServiceErr(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrPermissionDenied),
		errors.Is(err, service.ErrRoleNotAllowed),
		errors.Is(err, service.ErrCannotManageSelf):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrEmailTaken):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// ListUsers   GET /api/admin/users  atau  GET /api/super-admin/users
// admin       → hanya melihat role=user
// super_admin → melihat semua (dapat difilter dengan ?role=)
func (h *UserHandler) ListUsers(c fiber.Ctx) error {
	_, requesterRole := helper.ExtractRequesterInfo(c)

	var filter dto.UserFilter
	if err := c.Bind().Query(&filter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid query"})
	}

	result, err := h.userService.ListUsers(requesterRole, filter)
	if err != nil {
		return handleServiceErr(c, err)
	}
	return c.JSON(result)
}

// GetUser   GET /api/admin/users/:id  atau  GET /api/super-admin/users/:id
func (h *UserHandler) GetUser(c fiber.Ctx) error {
	requesterID, requesterRole := helper.ExtractRequesterInfo(c)
	tid, err := helper.ExtractTargetID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	user, err := h.userService.GetUser(requesterRole, requesterID, tid)
	if err != nil {
		return handleServiceErr(c, err)
	}
	return c.JSON(user)
}

// CreateUser   POST /api/admin/users  atau  POST /api/super-admin/users
// admin       → hanya dapat membuat pengguna dengan role=user
// super_admin → dapat membuat pengguna dengan role=admin atau role=user
func (h *UserHandler) CreateUser(c fiber.Ctx) error {
	_, requesterRole := helper.ExtractRequesterInfo(c)

	var req dto.CreateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	if err := middleware.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.userService.CreateUser(requesterRole, req)
	if err != nil {
		return handleServiceErr(c, err)
	}
	_ = h.auditLog.LogAction(nil, "USER_CREATED_BY_ADMIN", "Admin membuat pengguna baru: "+req.Email, c.IP(), c.Get("User-Agent"))
	return c.Status(fiber.StatusCreated).JSON(user)
}

// UpdateUser   PATCH /api/admin/users/:id  atau  PATCH /api/super-admin/users/:id
func (h *UserHandler) UpdateUser(c fiber.Ctx) error {
	requesterID, requesterRole := helper.ExtractRequesterInfo(c)
	tid, err := helper.ExtractTargetID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var req dto.UpdateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	if err := middleware.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.userService.UpdateUser(requesterRole, requesterID, tid, req)
	if err != nil {
		return handleServiceErr(c, err)
	}
	_ = h.auditLog.LogAction(&requesterID, "USER_UPDATED_BY_ADMIN", "Admin memperbarui data pengguna ID: "+tid.String(), c.IP(), c.Get("User-Agent"))
	return c.JSON(user)
}

// AssignRole   PATCH /api/super-admin/users/:id/role
// Hanya super_admin yang dapat mengatur ulang role (admin tidak dapat mengubah role pengguna menjadi admin).
func (h *UserHandler) AssignRole(c fiber.Ctx) error {
	requesterID, requesterRole := helper.ExtractRequesterInfo(c)
	tid, err := helper.ExtractTargetID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var req dto.AssignRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	if err := middleware.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.userService.AssignRole(requesterRole, requesterID, tid, req.Role)
	if err != nil {
		return handleServiceErr(c, err)
	}
	_ = h.auditLog.LogAction(&requesterID, "ROLE_ASSIGNED_BY_ADMIN", "Admin menetapkan role '"+string(req.Role)+"' ke pengguna ID: "+tid.String(), c.IP(), c.Get("User-Agent"))
	return c.JSON(user)
}

// SetActive   PATCH /api/admin/users/:id/active
func (h *UserHandler) SetActive(c fiber.Ctx) error {
	requesterID, requesterRole := helper.ExtractRequesterInfo(c)
	tid, err := helper.ExtractTargetID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var req dto.ToggleActiveRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	if err := h.userService.SetActive(requesterRole, requesterID, tid, req.IsActive); err != nil {
		return handleServiceErr(c, err)
	}
	status := "deactivated"
	action := "USER_DEACTIVATED"
	if req.IsActive {
		status = "activated"
		action = "USER_ACTIVATED"
	}
	_ = h.auditLog.LogAction(&requesterID, action, "Admin mengubah status aktif pengguna ID: "+tid.String()+" menjadi "+status, c.IP(), c.Get("User-Agent"))
	return c.JSON(fiber.Map{"message": "user " + status})
}

// DeleteUser   DELETE /api/admin/users/:id  atau  DELETE /api/super-admin/users/:id
func (h *UserHandler) DeleteUser(c fiber.Ctx) error {
	requesterID, requesterRole := helper.ExtractRequesterInfo(c)
	tid, err := helper.ExtractTargetID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	if err := h.userService.DeleteUser(requesterRole, requesterID, tid); err != nil {
		return handleServiceErr(c, err)
	}
	_ = h.auditLog.LogAction(&requesterID, "USER_DELETED_BY_ADMIN", "Admin menghapus (soft delete) pengguna ID: "+tid.String(), c.IP(), c.Get("User-Agent"))
	return c.JSON(fiber.Map{"message": "user deleted"})
}
