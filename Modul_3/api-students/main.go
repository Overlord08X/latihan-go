package main

import (
	"fmt"
	"log"
	"time"

	"api-students/app/repository"
	"api-students/config"
	"api-students/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	// 1. Muat konfigurasi dari .env
	cfg := config.Load()

	// 2. Buat connection pool ke PostgreSQL
	pool, err := database.Connect(cfg.DBDSN)
	if err != nil {
		log.Fatalf("[db] gagal terhubung: %v", err)
	}
	defer pool.Close()

	// 3. Jalankan migration SQL
	if err := database.RunMigrations(pool, "migrations"); err != nil {
		log.Fatalf("[db] gagal menjalankan migration: %v", err)
	}

	// 4. Buat repository dan handler melalui dependency injection
	repo := repository.NewStudentRepository(pool)
	h := NewStudentHandler(repo)

	// 5. Buat aplikasi Fiber
	app := fiber.New(fiber.Config{
		AppName: "Praktikum Backend Lanjut - Pertemuan 3",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			status := fiber.StatusInternalServerError
			pesan := "terjadi kesalahan pada server"
			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
				pesan = e.Message
			}
			return fail(c, status, pesan)
		},
	})

	// 6. Middleware global
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${locals:requestid} ${method} ${path} ${status} ${latency}\n",
	}))
	app.Use(cors.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("API Students — Praktikum Backend Lanjut Pertemuan 3")
	})

	api := app.Group("/api/v1")
	api.Get("/health", func(c *fiber.Ctx) error {
		return ok(c, "server berjalan", fiber.Map{"timestamp": time.Now()}, nil)
	})

	// 7. Route mahasiswa dengan middleware requireJSON per-grup
	s := api.Group("/students", requireJSON)
	s.Get("/", h.List)
	s.Get("/:id", h.Get)
	s.Post("/", h.Create)
	s.Put("/:id", h.Replace)
	s.Patch("/:id", h.Patch)
	s.Delete("/:id", h.Delete)

	// 8. Tangkap endpoint tidak dikenal
	app.Use(func(c *fiber.Ctx) error {
		return fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})

	fmt.Printf("Server berjalan di http://localhost:%s\n", cfg.AppPort)
	log.Fatal(app.Listen(":" + cfg.AppPort))
}
