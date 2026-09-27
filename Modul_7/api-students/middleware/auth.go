package middleware

import (
	"errors"
	"strings"

	"api-students/helper"
	"api-students/jwtutil"

	"github.com/gofiber/fiber/v2"
)

// RequireAuth adalah middleware yang memvalidasi JWT access token dari header Authorization.
// Bila valid, identitas pengguna (userID, username, role) disimpan di c.Locals.
func RequireAuth(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		c.Set("WWW-Authenticate", `Bearer realm="api"`)
		return helper.Unauthorized("token tidak ditemukan")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.Set("WWW-Authenticate", `Bearer realm="api"`)
		return helper.Unauthorized("format Authorization harus: Bearer <token>")
	}

	tokenStr := parts[1]
	claims, err := jwtutil.ParseAccessToken(tokenStr)
	if err != nil {
		c.Set("WWW-Authenticate", `Bearer realm="api"`)
		// Bedakan pesan kedaluwarsa vs tidak valid
		if errors.Is(err, jwtutil.ErrTokenExpired) {
			return helper.Unauthorized(jwtutil.ErrTokenExpired.Error())
		}
		return helper.Unauthorized(jwtutil.ErrTokenInvalid.Error())
	}

	// Simpan identitas ke Locals agar handler bisa menggunakannya
	c.Locals("userID", claims.Subject)
	c.Locals("username", claims.Username)
	c.Locals("role", claims.Role)

	return c.Next()
}
