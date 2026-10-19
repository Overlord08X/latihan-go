// Package config menyediakan fungsi untuk membaca konfigurasi aplikasi
// dari file .env dan environment variable.
package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config menyimpan seluruh konfigurasi yang dibutuhkan aplikasi.
type Config struct {
	AppPort string
	DBDSN   string
}

// Load membaca file .env (jika ada) dan mengembalikan Config yang sudah terisi.
// Nilai dari environment variable nyata akan menimpa nilai dari file .env,
// sehingga konfigurasi di lingkungan production tidak bergantung pada file .env.
func Load() *Config {
	// godotenv.Load tidak panic jika file .env tidak ada; aplikasi tetap bisa
	// berjalan selama environment variable sudah di-set dari luar.
	if err := godotenv.Load(); err != nil {
		log.Println("[config] file .env tidak ditemukan, menggunakan env var sistem")
	}

	return &Config{
		AppPort: getEnv("APP_PORT", "3000"),
		DBDSN:   getEnv("DB_DSN", "postgres://postgres:postgres@localhost:5432/praktikum_backend?sslmode=disable"),
	}
}

// GetEnv mengembalikan nilai environment variable atau nilai bawaan jika kosong.
func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// GetEnvInt membaca environment variable bertipe integer.
func GetEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

// getEnv mengembalikan nilai environment variable atau nilai bawaan jika kosong.
func getEnv(key, fallback string) string {
	return GetEnv(key, fallback)
}
