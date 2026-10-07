package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort     string
	DBDSN       string
	JWTSecret   string
	JWTExpHours int
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:171005@localhost:5432/siakad_db?sslmode=disable"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "siakad-mini-super-secret-jwt-key-2026-unair-pbe"
	}

	expHoursStr := os.Getenv("JWT_EXP_HOURS")
	expHours := 24
	if expHoursStr != "" {
		if parsed, err := strconv.Atoi(expHoursStr); err == nil && parsed > 0 {
			expHours = parsed
		}
	}

	return &Config{
		AppPort:     port,
		DBDSN:       dsn,
		JWTSecret:   secret,
		JWTExpHours: expHours,
	}
}
