package helper

import (
	"backend/internal/models"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// ExtractRequesterInfo retrieves the user UUID and highest role from the Fiber context locals.
func ExtractRequesterInfo(c fiber.Ctx) (uuid.UUID, models.Role) {
	idStr, _ := c.Locals("userID").(string)
	uid, _ := uuid.Parse(idStr)
	roleStr, _ := c.Locals("role").(string)
	return uid, models.Role(roleStr)
}

// ExtractTargetID parses the "id" parameter from the request path into a uuid.UUID.
func ExtractTargetID(c fiber.Ctx) (uuid.UUID, error) {
	return uuid.Parse(c.Params("id"))
}

// ExtractRoleNames collects the primary role and all role names from the JWT claims in the context locals.
func ExtractRoleNames(c fiber.Ctx) []string {
	var result []string
	if r, ok := c.Locals("role").(string); ok && r != "" {
		result = append(result, r)
	}
	if allRoles, ok := c.Locals("roles").([]interface{}); ok {
		seen := map[string]bool{}
		for _, item := range allRoles {
			if s, ok := item.(string); ok && !seen[s] {
				seen[s] = true
				result = append(result, s)
			}
		}
	}
	return result
}
