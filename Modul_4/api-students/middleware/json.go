// Package middleware menyediakan middleware Fiber untuk dipakai lintas route.
package middleware

import (
	"strings"

	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

// RequireJSON menolak request body dengan Content-Type selain application/json.
// Dipasang per-grup route (bukan global) agar tidak mempengaruhi endpoint lain.
func RequireJSON(c *fiber.Ctx) error {
	metodeBerbody := map[string]bool{
		fiber.MethodPost:  true,
		fiber.MethodPut:   true,
		fiber.MethodPatch: true,
	}
	if metodeBerbody[c.Method()] {
		ct := c.Get("Content-Type")
		if !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			return helper.Fail(c, fiber.StatusUnsupportedMediaType,
				"Content-Type harus application/json")
		}
	}
	return c.Next()
}
