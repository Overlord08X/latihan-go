package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitPool(dsn string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal parsing DB_DSN: %w", err)
	}

	cfg.MaxConns = 25
	cfg.MinConns = 5
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("gagal terhubung ke PostgreSQL: %w", err)
	}

	return pool, nil
}

func RunMigrations(pool *pgxpool.Pool, dir string) error {
	ctx := context.Background()

	// Tabel penanda riwayat migrasi
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return fmt.Errorf("gagal inisialisasi tabel schema_migrations: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return fmt.Errorf("gagal membaca folder migrasi: %w", err)
	}

	sort.Strings(files)

	for _, file := range files {
		base := filepath.Base(file)

		var exists bool
		err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", base).Scan(&exists)
		if err != nil {
			return fmt.Errorf("gagal memeriksa status migrasi %s: %w", base, err)
		}

		if exists {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("gagal membaca berkas migrasi %s: %w", file, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("gagal memulai transaksi migrasi %s: %w", base, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("gagal mengeksekusi migrasi %s: %w", base, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", base); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("gagal mencatat riwayat migrasi %s: %w", base, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("gagal commit transaksi migrasi %s: %w", base, err)
		}

		log.Printf("[Migration] Berhasil menerapkan: %s", base)
	}

	return nil
}
