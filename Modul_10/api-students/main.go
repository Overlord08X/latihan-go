package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-students/config"
	"api-students/database"
)

func main() {
	// 1. Muat konfigurasi .env
	cfg := config.Load()

	// 2. Hubungkan ke PostgreSQL
	db, err := database.Connect(cfg.DBDSN)
	if err != nil {
		slog.Error("gagal terhubung ke database PostgreSQL", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	// 2b. Hubungkan ke MongoDB
	mongoClient, err := database.NewMongoClient(context.Background())
	if err != nil {
		slog.Error("gagal terhubung ke MongoDB", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mongoClient.Disconnect(ctx)
	}()

	// 3. Jalankan migrasi PostgreSQL
	if err := database.RunMigrations(db, "migrations"); err != nil {
		slog.Error("gagal menjalankan migrasi", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 4. Rakit aplikasi (init logger, register route, dll)
	app := config.BuildApp(db, mongoClient)

	// 5. Siapkan channel untuk menangkap sinyal terminasi OS
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// 6. Jalankan server di goroutine terpisah
	port := ":" + cfg.AppPort
	go func() {
		slog.Info("memulai server", slog.String("port", port))
		if err := app.Listen(port); err != nil {
			slog.Error("gagal menjalankan server", slog.String("error", err.Error()))
		}
	}()

	// 7. Tunggu sinyal berhenti
	<-stop
	slog.Info("menerima sinyal berhenti, mematikan server...")

	// 8. Matikan server dengan rapi (graceful shutdown)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		slog.Error("gagal menutup server dengan rapi", slog.String("error", err.Error()))
	}
	slog.Info("server berhenti dengan rapi")
}
