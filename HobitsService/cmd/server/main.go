package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"

	"HobitsService/internal/app"
	"HobitsService/internal/config"
	"HobitsService/internal/infrastructure/database"
	httpserver "HobitsService/internal/infrastructure/http"
	"HobitsService/internal/logger"
	"HobitsService/internal/metrics"
)

func main() {
	cfg := config.MustLoad()

	// Инициализируем логирование
	if err := logger.Init(cfg.Env); err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Logger.Sync()

	// Инициализируем метрики
	if err := metrics.Init(); err != nil {
		log.Fatalf("failed to init metrics: %v", err)
	}

	ctx := context.Background()

	// Инициализируем базу данных
	db, err := database.New(ctx, &cfg.Postgres)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Запускаем миграции
	wd, err := os.Getwd()
	if err != nil {
		logger.Fatal("failed to get working directory", zap.Error(err))
	}

	migrationsPath := filepath.Join(wd, "migrations")
	if err := db.RunMigrations(migrationsPath); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	logger.Info("Successfully connected to database and ran migrations")

	// Инициализируем приложение (все зависимости)
	application := app.NewApp(db)
	defer application.Close()

	logger.Info("Application initialized successfully")

	// Запускаем HTTP сервер для метрик
	httpSrv := httpserver.NewServer(8080)
	if err := httpSrv.Start(); err != nil {
		logger.Fatal("failed to start HTTP server", zap.Error(err))
	}
	logger.Info("HTTP server started on port 8080 for /metrics and /health")

	// Запускаем gRPC сервер
	if err := application.GRPCServer.Start(); err != nil {
		logger.Fatal("failed to start gRPC server", zap.Error(err))
	}

	// Запускаем scheduler для автоматических задач
	application.Scheduler.Start()
	logger.Info("Scheduler started for streak checks and reminders")

	// Ожидаем сигнала завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Блокируемся здесь
	sig := <-sigChan
	logger.Info("received signal", zap.Any("signal", sig))

	// Грациозное завершение
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpSrv.Stop(ctx); err != nil {
		logger.Error("failed to stop HTTP server", zap.Error(err))
	}

	if err := application.GRPCServer.Stop(); err != nil {
		logger.Error("failed to stop gRPC server", zap.Error(err))
	}

	application.Close()
	logger.Info("Application shutdown completed")
}
