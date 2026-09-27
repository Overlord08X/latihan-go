package middleware

import (
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

// RequirePermission menolak request yang role-nya tidak memiliki permission tertentu.
func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("belum terautentikasi")
		}
		if !perms.Can(user.Role, permission) {
			return helper.Forbidden("role " + user.Role + " tidak memiliki hak " + permission)
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
			return helper.Unauthorized("belum terautentikasi")
		}
		if _, granted := allowed[user.Role]; !granted {
			return helper.Forbidden("role Anda tidak berhak mengakses endpoint ini")
		}
		return c.Next()
	}
}
