package auth

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const contextUserKey = "auth.user"

func RequireAuth(jwt *JWTService, users Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}

		claims, err := jwt.Validate(strings.TrimSpace(strings.TrimPrefix(header, prefix)), TokenTypeAccess)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}

		userID, err := strconv.Atoi(claims.Subject)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}
		user, err := users.GetByID(c.Context(), userID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if user == nil || !user.IsActive || user.Email != claims.Email || user.Role != claims.Role {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}

		c.Locals(contextUserKey, user)
		return c.Next()
	}
}

func RequireRole(roles ...Role) fiber.Handler {
	allowed := map[Role]bool{}
	for _, role := range roles {
		allowed[role] = true
	}
	return func(c *fiber.Ctx) error {
		user, ok := CurrentUser(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}
		if !allowed[user.Role] {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
		}
		return c.Next()
	}
}

func CurrentUser(c *fiber.Ctx) (*User, bool) {
	user, ok := c.Locals(contextUserKey).(*User)
	return user, ok
}
