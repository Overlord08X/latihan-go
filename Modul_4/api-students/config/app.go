// Package config menyatukan semua konfigurasi aplikasi (Fiber, Logger, Env).
package config

import (
	"time"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/helper"
	"api-students/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BuildApp merakit aplikasi Fiber, menginisialisasi middleware global,
// membuat service, dan mendaftarkan route.
func BuildApp(db *pgxpool.Pool) *fiber.App {
	// Inisialisasi Logger
	InitLogger()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Catch error default (404, 500, dll) dari Fiber
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return helper.Fail(c, code, err.Error())
		},
	})

	// Middleware Global
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(RequestLogger()) // Logger JSON buatan sendiri

	// Endpoint Health Check (dipindah dari main.go)
	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return helper.Success(c, fiber.StatusOK, "server berjalan", fiber.Map{
			"timestamp": time.Now().UTC(),
		})
	})


	// Inisiasi Layer
	studentRepo := repository.NewStudentRepository(db)
	studentSvc := service.NewStudentService(studentRepo)

	// Register Routes
	api := app.Group("/api/v1")
	route.RegisterStudent(api, studentSvc)

	// Fallback untuk route yang tidak ditemukan (404)
	app.Use(func(c *fiber.Ctx) error {
		return helper.Fail(c, fiber.StatusNotFound, "rute tidak ditemukan")
	})

	return app
}
