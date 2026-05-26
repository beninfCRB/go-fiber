package handler

import (
	"backend/internal/dto"
	"backend/internal/helper"
	"backend/internal/middleware"
	"backend/internal/models"
	"backend/internal/service"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService *service.AuthService
	auditLog    *service.AuditLogService
}

func NewAuthHandler(svc *service.AuthService, auditLog *service.AuditLogService) *AuthHandler {
	return &AuthHandler{authService: svc, auditLog: auditLog}
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	if err := middleware.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.authService.Register(req.Name, req.Email, req.Password); err != nil {
		_ = h.auditLog.LogAction(nil, "REGISTER_FAILED", "Pendaftaran gagal untuk email: "+req.Email+" - "+err.Error(), c.IP(), c.Get("User-Agent"))
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already registered"})
	}
	_ = h.auditLog.LogAction(nil, "USER_REGISTERED", "Pengguna baru berhasil mendaftar: "+req.Email, c.IP(), c.Get("User-Agent"))
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "registered successfully"})
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	if err := middleware.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	tokens, uid, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		_ = h.auditLog.LogAction(nil, "LOGIN_FAILED", "Percobaan masuk gagal untuk email: "+req.Email+" - "+err.Error(), c.IP(), c.Get("User-Agent"))
		if errors.Is(err, service.ErrUserInactive) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	_ = h.auditLog.LogAction(&uid, "LOGIN_SUCCESS", "Pengguna berhasil masuk ke sistem", c.IP(), c.Get("User-Agent"))
	return c.JSON(tokens)
}

func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	var req dto.RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	if err := middleware.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	tokens, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tokens)
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	var req dto.RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	if req.RefreshToken != "" {
		_ = h.authService.Logout(req.RefreshToken)
		_ = h.auditLog.LogAction(nil, "LOGOUT", "Pengguna keluar dari sistem dengan membatalkan Refresh Token", c.IP(), c.Get("User-Agent"))
	}
	return c.JSON(fiber.Map{"message": "logged out"})
}

func (h *AuthHandler) LogoutAll(c fiber.Ctx) error {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user id"})
	}
	if err := h.authService.LogoutAll(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	_ = h.auditLog.LogAction(&userID, "LOGOUT_ALL", "Pengguna keluar dari semua perangkat aktif", c.IP(), c.Get("User-Agent"))
	return c.JSON(fiber.Map{"message": "all sessions terminated"})
}

func (h *AuthHandler) Profile(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"userId": c.Locals("userID"),
		"name":   c.Locals("name"),
		"role":   c.Locals("role"),
		"roles":  c.Locals("roles"),
	})
}

// AdminListUsers — dapat diakses oleh admin + super_admin
func (h *AuthHandler) AdminListUsers(c fiber.Ctx) error {
	// TODO: injeksikan UserRepository dan kembalikan daftar pengguna terpaginasi
	return c.JSON(fiber.Map{"message": "admin: list users"})
}

// SuperAdminListRoles — hanya dapat diakses oleh super_admin
func (h *AuthHandler) SuperAdminListRoles(c fiber.Ctx) error {
	roles := []fiber.Map{
		{"name": models.RoleSuperAdmin, "description": "Full system access"},
		{"name": models.RoleAdmin, "description": "Tenant/organization admin"},
		{"name": models.RoleUser, "description": "Regular user"},
	}
	return c.JSON(roles)
}

// VerifyEmail menangani aktivasi akun pengguna melalui tautan email.
func (h *AuthHandler) VerifyEmail(c fiber.Ctx) error {
	token := c.Query("token")
	c.Set("Content-Type", "text/html")
	appURL := h.authService.GetAppURL()
	if token == "" {
		html := helper.GetEmailVerificationHTML(false, "Token verifikasi diperlukan", appURL)
		return c.Status(fiber.StatusBadRequest).SendString(html)
	}
	if err := h.authService.VerifyEmail(token); err != nil {
		_ = h.auditLog.LogAction(nil, "VERIFY_EMAIL_FAILED", "Verifikasi email gagal dengan token: "+token+" - "+err.Error(), c.IP(), c.Get("User-Agent"))
		html := helper.GetEmailVerificationHTML(false, err.Error(), appURL)
		return c.Status(fiber.StatusBadRequest).SendString(html)
	}
	_ = h.auditLog.LogAction(nil, "VERIFY_EMAIL_SUCCESS", "Email berhasil diverifikasi", c.IP(), c.Get("User-Agent"))
	html := helper.GetEmailVerificationHTML(true, "", appURL)
	return c.Status(fiber.StatusOK).SendString(html)
}

// ForgotPassword memproses permintaan tautan reset password.
func (h *AuthHandler) ForgotPassword(c fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	if err := middleware.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	
	if err := h.authService.ForgotPassword(req.Email); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	_ = h.auditLog.LogAction(nil, "FORGOT_PASSWORD_REQUEST", "Permintaan reset password diajukan untuk email: "+req.Email, c.IP(), c.Get("User-Agent"))
	
	return c.JSON(fiber.Map{"message": "jika email terdaftar, instruksi reset password telah dikirim"})
}

// ResetPassword memproses perubahan kata sandi baru menggunakan token reset yang valid.
func (h *AuthHandler) ResetPassword(c fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	if err := middleware.Validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	
	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		_ = h.auditLog.LogAction(nil, "RESET_PASSWORD_FAILED", "Reset password gagal dengan token: "+req.Token+" - "+err.Error(), c.IP(), c.Get("User-Agent"))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	_ = h.auditLog.LogAction(nil, "RESET_PASSWORD_SUCCESS", "Kata sandi berhasil diperbarui menggunakan token reset", c.IP(), c.Get("User-Agent"))
	
	return c.JSON(fiber.Map{"message": "kata sandi berhasil diperbarui, silakan login"})
}
