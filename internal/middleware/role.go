package middleware

import (
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
)

func RoleGuard(allowed ...models.Role) fiber.Handler {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, r := range allowed {
		allowedSet[string(r)] = struct{}{}
	}

	return func(c fiber.Ctx) error {
		if primaryRole, ok := c.Locals("role").(string); ok {
			if _, ok := allowedSet[primaryRole]; ok {
				return c.Next()
			}
		}

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
