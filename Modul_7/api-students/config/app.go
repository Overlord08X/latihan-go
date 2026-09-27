// Package config menyatukan semua konfigurasi aplikasi (Fiber, Logger, Env).
package config

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"api-students/app/model"
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

// newErrorHandler adalah SATU-SATUNYA tempat error berubah menjadi
// response HTTP di seluruh aplikasi (Langkah 2).
//
// Perbaikan Bug Compiler #1: Mengakses error asli via appErr.Cause() / Unwrap(), bukan unexported field cause.
// Perbaikan Bug Perilaku #2: Kondisi status >= 500 dicatat sebagai Error ("request_failed"),
// sedangkan status < 500 (4xx) dicatat sebagai Warn ("request_rejected").
func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		requestID := helper.RequestID(c)
		var appErr *helper.AppError

		switch {
		case errors.As(err, &appErr):
			// Kegagalan yang sudah kita rencanakan.
		case errors.Is(err, fiber.ErrRequestEntityTooLarge):
			appErr = &helper.AppError{
				Status:  fiber.StatusRequestEntityTooLarge,
				Code:    "PAYLOAD_TOO_LARGE",
				Message: "ukuran body melebihi batas yang diizinkan",
			}
		default:
			// Kegagalan yang tidak kita duga.
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				appErr = &helper.AppError{
					Status:  fiberErr.Code,
					Code:    "HTTP_ERROR",
					Message: fiberErr.Message,
				}
			} else {
				appErr = helper.Internal(err)
			}
		}

		// Hanya kegagalan sisi server (5xx) yang dicatat sebagai Error.
		// Kegagalan 4xx adalah kesalahan pemakai API, dicatat sebagai Warn.
		if appErr.Status >= fiber.StatusInternalServerError {
			errDetail := ""
			if appErr.Cause() != nil {
				errDetail = appErr.Cause().Error()
			} else {
				errDetail = appErr.Error()
			}
			logger.Error("request_failed",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status),
				slog.String("error", errDetail))
		} else {
			logger.Warn("request_rejected",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status))
		}

		return c.Status(appErr.Status).JSON(model.ErrorResponse{
			Success:   false,
			Code:      appErr.Code,
			Message:   appErr.Message,
			Fields:    appErr.Fields,
			RequestID: requestID,
		})
	}
}

// BuildApp merakit aplikasi Fiber, menginisialisasi middleware global,
// memuat pemetaan RBAC ke memori, membuat service, dan mendaftarkan route.
func BuildApp(db *pgxpool.Pool) *fiber.App {
	// 1. Inisialisasi Logger
	InitLogger()

	// 2. Inisialisasi JWT Secret — aplikasi panic jika JWT_SECRET kosong
	jwtutil.MustLoadJWTSecret()

	app := fiber.New(fiber.Config{
		// Batasi ukuran body request untuk mencegah payload raksasa menghabiskan memori
		BodyLimit:    1 * 1024 * 1024, // 1 MB
		ErrorHandler: newErrorHandler(slog.Default()),
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

	// 5. Inisiasi Repository & Memuat RBAC PermissionSet ke Memori
	roleRepo := repository.NewRoleRepository(db)
	rawPermissions, err := roleRepo.LoadPermissions(context.Background())
	if err != nil {
		slog.Error("gagal memuat permission dari database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	permissions := helper.NewPermissionSet(rawPermissions)
	slog.Info("permission dimuat", slog.Any("roles", permissions.KnownRoles()))

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	nilaiRepo := repository.NewNilaiRepository(db)

	// 6. Inisiasi Service Layer dengan Dependency Injection
	authSvc := service.NewAuthService(userRepo, tokenRepo, permissions)
	userSvc := service.NewUserService(userRepo, permissions)
	studentSvc := service.NewStudentService(studentRepo, permissions)
	nilaiSvc := service.NewNilaiService(nilaiRepo)

	// 7. Rate limiter untuk endpoint login — mencegah brute force
	loginLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 30 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			c.Set("Retry-After", "30")
			return helper.TooManyRequests("terlalu banyak percobaan login, tunggu 30 detik")
		},
	})

	// 8. Register Routes
	api := app.Group("/api/v1")
	route.RegisterAuth(api, authSvc, loginLimiter)
	route.RegisterUser(api, userSvc, permissions)
	route.RegisterStudent(api, studentSvc, permissions)
	route.RegisterNilai(api, nilaiSvc)

	// 9. Fallback untuk route yang tidak ditemukan (404) melalui ErrorHandler terpusat
	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})

	return app
}
