package middleware

import (
	"github.com/gofiber/fiber/v2"
	"siakad-mini/helper"
)

// RequirePermission menolak request yang role-nya tidak memiliki permission tertentu.
func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.ErrorResponsePayload(c, fiber.StatusUnauthorized, "belum terautentikasi", nil)
		}
		if !perms.Can(user.Role, permission) {
			return helper.ErrorResponsePayload(c, fiber.StatusForbidden, "role " + user.Role + " tidak memiliki hak " + permission, nil)
		}
		return c.Next()
	}
}

// RequireRole memeriksa nama role secara langsung.
func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.ErrorResponsePayload(c, fiber.StatusUnauthorized, "belum terautentikasi", nil)
		}
		if _, granted := allowed[user.Role]; !granted {
			return helper.ErrorResponsePayload(c, fiber.StatusForbidden, "role Anda tidak berhak mengakses endpoint ini", nil)
		}
		return c.Next()
	}
}
