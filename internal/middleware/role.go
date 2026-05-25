package middleware

import (
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
)

// RoleGuard restricts access to users that have ANY of the specified roles.
// Requires JWTMiddleware to run first.
//
// Usage:
//
//	api.Get("/admin", middleware.RoleGuard(models.RoleAdmin, models.RoleSuperAdmin))
func RoleGuard(allowed ...models.Role) fiber.Handler {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, r := range allowed {
		allowedSet[string(r)] = struct{}{}
	}

	return func(c fiber.Ctx) error {
		// Check against the primary role (string) first — fast path
		if primaryRole, ok := c.Locals("role").(string); ok {
			if _, ok := allowedSet[primaryRole]; ok {
				return c.Next()
			}
		}

		// Check against all roles ([]interface{} from JWT MapClaims)
		if allRoles, ok := c.Locals("roles").([]interface{}); ok {
			for _, r := range allRoles {
				if roleName, ok := r.(string); ok {
					if _, ok := allowedSet[roleName]; ok {
						return c.Next()
					}
				}
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "forbidden: insufficient permissions",
		})
	}
}
