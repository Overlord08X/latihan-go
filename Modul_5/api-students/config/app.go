// Package config menyatukan semua konfigurasi aplikasi (Fiber, Logger, Env).
package config

import (
	"time"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/helper"
	"api-students/jwtutil"
	"api-students/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BuildApp merakit aplikasi Fiber, menginisialisasi middleware global,
// membuat service, dan mendaftarkan route.
func BuildApp(db *pgxpool.Pool) *fiber.App {
	// 1. Inisialisasi Logger
	InitLogger()

	// 2. Inisialisasi JWT Secret — aplikasi panic jika JWT_SECRET kosong
	jwtutil.MustLoadJWTSecret()

	app := fiber.New(fiber.Config{
		// Batasi ukuran body request untuk mencegah payload raksasa menghabiskan memori
		BodyLimit: 1 * 1024 * 1024, // 1 MB
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return helper.Fail(c, code, err.Error())
		},
	})

	// 3. Middleware Global
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(RequestLogger())

	// 4. Endpoint Health Check — tetap publik, tidak memerlukan token
	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return helper.Success(c, fiber.StatusOK, "server berjalan", fiber.Map{
			"timestamp": time.Now().UTC(),
		})
	})

	// 5. Inisiasi semua Layer
	studentRepo := repository.NewStudentRepository(db)
	studentSvc := service.NewStudentService(studentRepo)

	nilaiRepo := repository.NewNilaiRepository(db)
	nilaiSvc := service.NewNilaiService(nilaiRepo)

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	authSvc := service.NewAuthService(userRepo, tokenRepo)

	// 6. Rate limiter untuk endpoint login — mencegah brute force
	// Max 5 percobaan per IP per 30 detik; percobaan ke-6+ → 429 + Retry-After
	loginLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 30 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			c.Set("Retry-After", "30")
			return helper.Fail(c, fiber.StatusTooManyRequests,
				"terlalu banyak percobaan login, tunggu 30 detik")
		},
	})

	// 7. Register Routes
	api := app.Group("/api/v1")
	// loginLimiter diteruskan ke RegisterAuth agar diterapkan sebelum handler login
	route.RegisterAuth(api, authSvc, loginLimiter)
	route.RegisterStudent(api, studentSvc)
	route.RegisterNilai(api, nilaiSvc)

	// 8. Fallback untuk route yang tidak ditemukan (404)
	app.Use(func(c *fiber.Ctx) error {
		return helper.Fail(c, fiber.StatusNotFound, "rute tidak ditemukan")
	})

	return app
}
