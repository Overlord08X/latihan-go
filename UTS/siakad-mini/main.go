package main

import (
	"fmt"
	"log"
	"time"

	"siakad-mini/app/repository"
	"siakad-mini/app/service"
	"siakad-mini/config"
	"siakad-mini/database"
	"siakad-mini/middleware"
	"siakad-mini/route"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// 1. Muat konfigurasi
	cfg := config.LoadConfig()

	// 2. Hubungkan ke database PostgreSQL
	pool, err := database.InitPool(cfg.DBDSN)
	if err != nil {
		log.Fatalf("Fatal: Gagal menghubungkan ke database: %v", err)
	}
	defer pool.Close()
	log.Println("[Database] Berhasil terhubung ke PostgreSQL")

	// 3. Jalankan migrasi otomatis dan seeder
	if err := database.RunMigrations(pool, "./migrations"); err != nil {
		log.Fatalf("Fatal: Gagal menjalankan migrasi database: %v", err)
	}

	// 4. Inisialisasi FailureTracker (Rate limiter: maks 5 kegagalan login per 1 menit)
	failureTracker := middleware.NewFailureTracker(5, 1*time.Minute)

	// 5. Inisialisasi Repository
	userRepo := repository.NewUserRepository(pool)
	studentRepo := repository.NewStudentRepository(pool)
	courseRepo := repository.NewCourseRepository(pool)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)

	// 6. Inisialisasi Service
	authService := service.NewAuthService(userRepo, studentRepo, cfg.JWTSecret, cfg.JWTExpHours, failureTracker)
	studentService := service.NewStudentService(pool, studentRepo, userRepo, enrollmentRepo)
	courseService := service.NewCourseService(courseRepo)
	enrollmentService := service.NewEnrollmentService(pool, enrollmentRepo, courseRepo, studentRepo)

	// 7. Bangun Fiber App
	app := config.NewApp()

	// 8. Daftarkan 10 Endpoint di bawah grup /api/v1
	api := app.Group("/api/v1")
	route.RegisterAuth(api, authService, cfg.JWTSecret)
	route.RegisterStudent(api, studentService, cfg.JWTSecret)
	route.RegisterCourse(api, courseService, cfg.JWTSecret)
	route.RegisterEnrollment(api, enrollmentService, cfg.JWTSecret)

	// Root health check endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"app":     "SIAKAD Mini API",
			"status":  "running",
			"version": "1.0.0",
		})
	})

	// 9. Jalankan server
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("[Server] Menjalankan server pada port %s...", cfg.AppPort)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Fatal: Gagal menjalankan server: %v", err)
	}
}
