package handler

import (
	"backend/internal/dto"
	"backend/internal/helper"
	"backend/internal/middleware"
	"backend/internal/service"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type MenuHandler struct {
	menuService *service.MenuService
	auditLog    *service.AuditLogService
}

func NewMenuHandler(svc *service.MenuService, auditLog *service.AuditLogService) *MenuHandler {
	return &MenuHandler{menuService: svc, auditLog: auditLog}
}

func handleMenuErr(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrMenuNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrMenuKeyTaken):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, service.ErrRoleNotFound):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}

// GetMyMenus   GET /api/menu
// Mengembalikan pohon navigasi menu untuk pengguna yang terautentikasi saat ini.
// Frontend memanggil endpoint ini sekali setelah login untuk menyusun sidebar.
func (h *MenuHandler) GetMyMenus(c fiber.Ctx) error {
	// role dapat berupa string tunggal atau []interface{} tergantung pada token claims
	roleNames := helper.ExtractRoleNames(c)
	menus, err := h.menuService.GetUserMenuTree(roleNames)
	if err != nil {
		return handleMenuErr(c, err)
	}
	return c.JSON(fiber.Map{"menus": menus})
}

// ListMenus   GET /api/super-admin/menus
// Mengembalikan daftar datar semua menu beserta role yang diizinkan.
func (h *MenuHandler) ListMenus(c fiber.Ctx) error {
	menus, err := h.menuService.ListAllMenus()
	if err != nil {
		return handleMenuErr(c, err)
	}
	return c.JSON(fiber.Map{"menus": menus})
}

// GetMenuTree   GET /api/super-admin/menus/tree
// Mengembalikan struktur pohon hierarki menu lengkap untuk super_admin.
func (h *MenuHandler) GetMenuTree(c fiber.Ctx) error {
	tree, err := h.menuService.GetMenuTree()
	if err != nil {
		return handleMenuErr(c, err)
	}
	return c.JSON(fiber.Map{"menus": tree})
}

// GetMenu   GET /api/super-admin/menus/:id
func (h *MenuHandler) GetMenu(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	menu, err := h.menuService.GetMenu(id)
	if err != nil {
		return handleMenuErr(c, err)
	}
	return c.JSON(menu)
}

// CreateMenu   POST /api/super-admin/menus
func (h *MenuHandler) CreateMenu(c fiber.Ctx) error {
	requesterID, _ := helper.ExtractRequesterInfo(c)
	var req dto.CreateMenuRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	if err := middleware.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	menu, err := h.menuService.CreateMenu(req)
	if err != nil {
		return handleMenuErr(c, err)
	}
	_ = h.auditLog.LogAction(&requesterID, "MENU_CREATED", "Super Admin membuat item menu baru: "+req.Name+" (key: "+req.Key+")", c.IP(), c.Get("User-Agent"))
	return c.Status(fiber.StatusCreated).JSON(menu)
}

// UpdateMenu   PATCH /api/super-admin/menus/:id
func (h *MenuHandler) UpdateMenu(c fiber.Ctx) error {
	requesterID, _ := helper.ExtractRequesterInfo(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	var req dto.UpdateMenuRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	if err := middleware.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	menu, err := h.menuService.UpdateMenu(id, req)
	if err != nil {
		return handleMenuErr(c, err)
	}
	_ = h.auditLog.LogAction(&requesterID, "MENU_UPDATED", "Super Admin memperbarui data item menu ID: "+id.String(), c.IP(), c.Get("User-Agent"))
	return c.JSON(menu)
}

// AssignMenuRoles   PUT /api/super-admin/menus/:id/roles
// Mengganti daftar role yang diizinkan untuk melihat menu ini.
func (h *MenuHandler) AssignMenuRoles(c fiber.Ctx) error {
	requesterID, _ := helper.ExtractRequesterInfo(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	var req dto.AssignMenuRolesRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	if err := middleware.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	menu, err := h.menuService.AssignMenuRoles(id, req.RoleKeys)
	if err != nil {
		return handleMenuErr(c, err)
	}
	_ = h.auditLog.LogAction(&requesterID, "MENU_ROLES_ASSIGNED", "Super Admin mengubah hak akses role pada menu ID: "+id.String(), c.IP(), c.Get("User-Agent"))
	return c.JSON(menu)
}

// DeleteMenu   DELETE /api/super-admin/menus/:id
func (h *MenuHandler) DeleteMenu(c fiber.Ctx) error {
	requesterID, _ := helper.ExtractRequesterInfo(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.menuService.DeleteMenu(id); err != nil {
		return handleMenuErr(c, err)
	}
	_ = h.auditLog.LogAction(&requesterID, "MENU_DELETED", "Super Admin menghapus item menu ID: "+id.String(), c.IP(), c.Get("User-Agent"))
	return c.JSON(fiber.Map{"message": "menu deleted"})
}

// extractRoleNames reads all role names from JWT locals.
