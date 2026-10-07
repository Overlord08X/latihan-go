package middleware

import (
	"strings"

	"siakad-mini/helper"
	"siakad-mini/jwtutil"

	"github.com/gofiber/fiber/v2"
)

func RequireAuth(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helper.Unauthorized("Token otentikasi tidak ditemukan")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return helper.Unauthorized("Format header Authorization harus: Bearer <token>")
		}

		claims, err := jwtutil.ParseToken(parts[1], jwtSecret)
		if err != nil {
			return helper.Unauthorized(err.Error())
		}

		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		c.Locals("studentID", claims.StudentID)

		return c.Next()
	}
}

func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok || userRole == "" {
			return helper.Forbidden("Role pengguna tidak teridentifikasi")
		}

		for _, role := range allowedRoles {
			if strings.EqualFold(userRole, role) {
				return c.Next()
			}
		}

		return helper.Forbidden("Akses ditolak: role Anda tidak memiliki izin untuk endpoint ini")
	}
}
