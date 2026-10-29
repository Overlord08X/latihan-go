package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-students/config"
	"api-students/database"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	config.LoadEnv()
	cfg := config.Load()
	logger := slog.Default()

	role := flag.String("role", config.GetEnv("APP_ROLE", "all"), "api | worker | all")
	flag.Parse()

	runAPI := *role == "api" || *role == "all"
	runWorker := *role == "worker" || *role == "all"
	if !runAPI && !runWorker {
		logger.Error("role tidak dikenal", slog.String("role", *role))
		os.Exit(1)
	}

	db, err := database.Connect(cfg.DBDSN)
	if err != nil {
		logger.Error("gagal terhubung ke database PostgreSQL", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	if err := database.RunMigrations(db, "migrations"); err != nil {
		logger.Error("gagal menjalankan migrasi", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var app *fiber.App
	var mongoClient *mongo.Client

	if runAPI {
		var err error
		mongoClient, err = database.NewMongoClient(context.Background())
		if err != nil {
			logger.Error("gagal terhubung ke MongoDB", slog.String("error", err.Error()))
			os.Exit(1)
		}

		app = config.BuildApp(db, mongoClient)
		port := ":" + cfg.AppPort
		go func() {
			logger.Info("HTTP berjalan", slog.String("port", port))
			if err := app.Listen(port); err != nil {
				logger.Error("server HTTP berhenti", slog.String("error", err.Error()))
			}
		}()
	}

	var stopWorker func()
	if runWorker {
		stopWorker = startWorker(ctx, logger, db)
		if stopWorker == nil {
			os.Exit(1)
		}
	}

	logger.Info("aplikasi berjalan", slog.String("role", *role))
	<-ctx.Done()
	logger.Info("menerima sinyal berhenti, mematikan layanan...")

	if app != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			logger.Error("shutdown HTTP tidak rapi", slog.String("error", err.Error()))
		}
		cancel()
	}

	if stopWorker != nil {
		stopWorker()
	}

	if mongoClient != nil {
		discCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = mongoClient.Disconnect(discCtx)
		cancel()
	}

	logger.Info("aplikasi berhenti dengan rapi")
}
