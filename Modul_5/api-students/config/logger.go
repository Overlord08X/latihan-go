package config

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger mengatur slog untuk menulis ke stdout dan file log
// dengan rotasi.
func InitLogger() {
	// Buat folder logs jika belum ada
	_ = os.MkdirAll("logs", 0755)

	// Rotasi log dengan lumberjack
	fileLogger := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   true,
	}

	// Tulis ke file dan console sekaligus
	multiWriter := io.MultiWriter(os.Stdout, fileLogger)

	// Format JSON
	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// RequestLogger adalah middleware Fiber untuk mencatat setiap request.
// Setiap baris log akan berformat JSON memuat request_id, method, dll.
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Lanjutkan eksekusi
		err := c.Next()

		duration := time.Since(start)

		// Kumpulkan data log
		reqID := c.Get("X-Request-Id")
		if reqID == "" {
			reqID = "N/A"
		}

		slog.Info("Request diproses",
			slog.String("request_id", reqID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.String("duration", duration.String()),
			slog.String("ip", c.IP()),
		)

		return err
	}
}
