// Package config menyatukan semua konfigurasi aplikasi (Fiber, Logger, Env).
package config

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"api-students/helper"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger mengatur structured logging dengan format JSON.
func InitLogger() {
	logDir := "logs"
	_ = os.MkdirAll(logDir, 0755)

	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "app.log"),
		MaxSize:    10, // Megabytes
		MaxBackups: 5,
		MaxAge:     30, // Hari
		Compress:   true,
	}

	multiWriter := io.MultiWriter(os.Stdout, fileWriter)

	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// RequestLogger adalah middleware Fiber untuk mencatat setiap request HTTP (Langkah 3).
// Perbaikan Bug Perilaku #3: Menggunakan variabel status yang telah diperbaiki saat err != nil,
// bukan c.Response().StatusCode() yang masih bernilai bawaan (200) sebelum ErrorHandler berjalan.
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		requestID, _ := c.Locals("requestid").(string)
		if requestID == "" {
			requestID = c.Get("X-Request-Id")
		}

		status := c.Response().StatusCode()
		if err != nil {
			var appErr *helper.AppError
			if errors.As(err, &appErr) {
				status = appErr.Status
			} else {
				var fiberErr *fiber.Error
				if errors.As(err, &fiberErr) {
					status = fiberErr.Code
				} else {
					status = fiber.StatusInternalServerError
				}
			}
		}

		attrs := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		}

		// Identitas ikut dicatat bila request sudah melewati RequireAuth.
		if user, ok := helper.CurrentUser(c); ok {
			attrs = append(attrs,
				slog.Int("user_id", user.UserID),
				slog.String("role", user.Role),
			)
		}

		slog.Info("http_request", attrs...)

		return err
	}
}
