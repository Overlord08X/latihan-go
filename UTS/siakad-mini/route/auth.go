package route

import (
	"siakad-mini/app/model"
	"siakad-mini/app/service"
	"siakad-mini/helper"
	"siakad-mini/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterAuth(r fiber.Router, authService *service.AuthService, jwtSecret string) {
	auth := r.Group("/auth")

	// Endpoint 1: POST /api/v1/auth/login (Publik)
	auth.Post("/login", func(c *fiber.Ctx) error {
		var req model.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.UnprocessableEntity("Format JSON tidak valid", map[string][]string{
				"body": {"Body request harus berformat JSON yang valid"},
			})
		}

		res, err := authService.Login(c.Context(), c.IP(), req)
		if err != nil {
			return err
		}

		return helper.Success(c, fiber.StatusOK, "Login berhasil", res)
	})

	// Endpoint 2: GET /api/v1/auth/me (Semua role terotentikasi)
	auth.Get("/me", middleware.RequireAuth(jwtSecret), func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(int64)
		role := c.Locals("role").(string)

		res, err := authService.Me(c.Context(), userID, role)
		if err != nil {
			return err
		}

		return helper.Success(c, fiber.StatusOK, "Profil pengguna berhasil diambil", res)
	})
}
