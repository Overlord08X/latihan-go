// Package database menyediakan connection pool ke PostgreSQL menggunakan pgxpool.
package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect membuat connection pool ke PostgreSQL dan menjalankan migration awal.
// Menggunakan pgxpool agar koneksi dapat digunakan secara concurrent tanpa
// membuka koneksi baru setiap request.
func Connect(dsn string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pool: %w", err)
	}

	// Ping untuk memastikan koneksi berhasil sebelum server mulai menerima request.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("gagal ping database: %w", err)
	}

	log.Println("[db] berhasil terhubung ke PostgreSQL")
	return pool, nil
}

// RunMigrations membaca file SQL dari folder migrations/ dan menjalankannya.
// Pendekatan manual ini dipilih agar mahasiswa memahami cara kerja migration
// sebelum menggunakan library migration seperti golang-migrate.
func RunMigrations(pool *pgxpool.Pool, migrationDir string) error {
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("gagal membaca direktori migration: %w", err)
	}

	ctx := context.Background()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := migrationDir + "/" + entry.Name()
		sql, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("gagal membaca %s: %w", path, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("gagal menjalankan %s: %w", path, err)
		}
		log.Printf("[db] migration dijalankan: %s", entry.Name())
	}
	return nil
}
