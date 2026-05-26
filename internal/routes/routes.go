package routes

import (
	"backend/internal/handler"
	"backend/internal/helper"
	"backend/internal/middleware"
	"backend/internal/models"
	"crypto/rsa"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func Register(app *fiber.App, h *handler.Handlers, jwtPublicKey *rsa.PublicKey) {
	// Swagger UI & Specification
	app.Get("/swagger", helper.ServeSwaggerUI)
	app.Get("/swagger.json", helper.ServeSwaggerJSON)

	registerAuth(app, h.Auth)
	registerAPI(app, h, jwtPublicKey)
}

func registerAuth(app *fiber.App, h *handler.AuthHandler) {
	authLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "terlalu banyak percobaan, silakan coba lagi dalam satu menit",
			})
		},
	})

	auth := app.Group("/auth")
	auth.Post("/register", authLimiter, h.Register)
	auth.Post("/login", authLimiter, h.Login)
	auth.Get("/verify-email", h.VerifyEmail)
	auth.Post("/forgot-password", authLimiter, h.ForgotPassword)
	auth.Post("/reset-password", authLimiter, h.ResetPassword)
	auth.Post("/refresh", h.Refresh)
	auth.Post("/logout", h.Logout)
}

func registerAPI(app *fiber.App, h *handler.Handlers, jwtPublicKey *rsa.PublicKey) {
	api := app.Group("/api", middleware.JWTMiddleware(jwtPublicKey))

	api.Get("/profile", h.Auth.Profile)
	api.Post("/logout-all", h.Auth.LogoutAll)
	api.Get("/menu", h.Menu.GetMyMenus)

	// ── Admin + Super Admin ───────────────────────────────────────────────────
	registerAdminRoutes(api, h)

	// ── Khusus Super Admin ────────────────────────────────────────────────────
	registerSuperAdminRoutes(api, h)
}

// registerAdminRoutes — dapat diakses oleh admin DAN super_admin.
//
// Menu: User Management
//   GET    /api/admin/users          → list users
//   POST   /api/admin/users          → create user (role=user only for admin)
//   GET    /api/admin/users/:id      → user detail
//   PATCH  /api/admin/users/:id      → update user info
//   PATCH  /api/admin/users/:id/active → activate/deactivate
//   DELETE /api/admin/users/:id      → soft delete
func registerAdminRoutes(router fiber.Router, h *handler.Handlers) {
	admin := router.Group("/admin",
		middleware.RoleGuard(models.RoleAdmin, models.RoleSuperAdmin),
	)
	admin.Get("/users", h.User.ListUsers)
	admin.Post("/users", h.User.CreateUser)
	admin.Get("/users/:id", h.User.GetUser)
	admin.Patch("/users/:id", h.User.UpdateUser)
	admin.Patch("/users/:id/active", h.User.SetActive)
	admin.Delete("/users/:id", h.User.DeleteUser)
}

// registerSuperAdminRoutes — hanya dapat diakses oleh super_admin.
//
// Menu: User Management (extended)
//   PATCH /api/super-admin/users/:id/role → change user role
//
// Menu: Menu Management
//   GET    /api/super-admin/menus          → flat list with role assignments
//   GET    /api/super-admin/menus/tree     → nested tree view
//   GET    /api/super-admin/menus/:id      → single menu detail
//   POST   /api/super-admin/menus          → create menu
//   PATCH  /api/super-admin/menus/:id      → update menu
//   PUT    /api/super-admin/menus/:id/roles → assign roles to menu
//   DELETE /api/super-admin/menus/:id      → delete menu
//
// Menu: Role Management
//   GET /api/super-admin/roles → list all roles
func registerSuperAdminRoutes(router fiber.Router, h *handler.Handlers) {
	sa := router.Group("/super-admin",
		middleware.RoleGuard(models.RoleSuperAdmin),
	)

	// User role assignment
	sa.Patch("/users/:id/role", h.User.AssignRole)

	// Role catalog
	sa.Get("/roles", h.Auth.SuperAdminListRoles)

	// Menu management
	sa.Get("/menus", h.Menu.ListMenus)
	sa.Get("/menus/tree", h.Menu.GetMenuTree)
	sa.Post("/menus", h.Menu.CreateMenu)
	sa.Get("/menus/:id", h.Menu.GetMenu)
	sa.Patch("/menus/:id", h.Menu.UpdateMenu)
	sa.Put("/menus/:id/roles", h.Menu.AssignMenuRoles)
	sa.Delete("/menus/:id", h.Menu.DeleteMenu)
}
