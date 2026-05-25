package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware validates the Bearer token and populates Locals:
//   - "userID" float64  — user primary key
//   - "name"   string   — user display name
//   - "role"   string   — highest-privilege role (e.g. "super_admin")
//   - "roles"  []string — all roles the user has
func JWTMiddleware(secret []byte) fiber.Handler {
	return func(c fiber.Ctx) error {
		auth := c.Get("Authorization")
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		if tokenStr == "" || tokenStr == auth {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing or malformed token",
			})
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return secret, nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token claims",
			})
		}

		c.Locals("userID", claims["sub"])
		c.Locals("name", claims["name"])
		c.Locals("role", claims["role"])    // highest role (string)
		c.Locals("roles", claims["roles"]) // all roles ([]interface{})
		return c.Next()
	}
}
