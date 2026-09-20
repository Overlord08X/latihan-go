package middleware

import (
	"api-students/jwtutil"
	"api-students/helper"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// RequireAuth adalah middleware yang memvalidasi JWT access token dari header Authorization.
// Bila valid, identitas pengguna (userID, username, role) disimpan di c.Locals.
//
// Membedakan "kedaluwarsa" vs "tidak valid" aman dilakukan pada token —
// berbeda dengan login, di mana kita tidak boleh membedakan "username tidak ada"
// dan "password salah" (user enumeration risk).
// Token bukan rahasia dari sisi validasi: fakta bahwa ia kedaluwarsa bukan
// informasi yang membantu penyerang.
func RequireAuth(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		c.Set("WWW-Authenticate", `Bearer realm="api"`)
		return helper.Fail(c, fiber.StatusUnauthorized, "token tidak ditemukan")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.Set("WWW-Authenticate", `Bearer realm="api"`)
		return helper.Fail(c, fiber.StatusUnauthorized, "format Authorization harus: Bearer <token>")
	}

	tokenStr := parts[1]
	claims, err := jwtutil.ParseAccessToken(tokenStr)
	if err != nil {
		c.Set("WWW-Authenticate", `Bearer realm="api"`)
		// Bedakan pesan kedaluwarsa vs tidak valid
		if errors.Is(err, jwtutil.ErrTokenExpired) {
			return helper.Fail(c, fiber.StatusUnauthorized, jwtutil.ErrTokenExpired.Error())
		}
		return helper.Fail(c, fiber.StatusUnauthorized, jwtutil.ErrTokenInvalid.Error())
	}

	// Simpan identitas ke Locals agar handler bisa menggunakannya
	c.Locals("userID", claims.Subject)
	c.Locals("username", claims.Username)
	c.Locals("role", claims.Role)

	return c.Next()
}
