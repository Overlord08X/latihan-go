package route

import (
	"api-students/app/service"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
)

// RegisterAuth mendaftarkan endpoint autentikasi.
// loginLimiter adalah middleware rate limiter yang diterapkan KHUSUS pada POST /auth/login.
func RegisterAuth(api fiber.Router, svc *service.AuthService, loginLimiter fiber.Handler) {
	auth := api.Group("/auth", middleware.RequireJSON)

	// Endpoint publik — tidak perlu token
	auth.Post("/register", svc.Register)
	// Rate limiter diterapkan SEBELUM handler login — urutan penting di Fiber
	auth.Post("/login", loginLimiter, svc.Login)
	auth.Post("/refresh", svc.Refresh)
	auth.Post("/logout", svc.Logout)

	// GET /auth/me — butuh token yang valid
	auth.Get("/me", middleware.RequireAuth, svc.Me)
}
