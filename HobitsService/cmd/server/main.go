package main

import (
	"HobitsService/internal/config"
	"HobitsService/internal/infrastructure/database"
	"context"
	"log"
	"os"
	"path/filepath"
)

func main() {
	cfg := config.MustLoad()

	log.Printf("Config loaded: ENV=%s, DB_HOST=%s, DB_PORT=%d, DB_NAME=%s",
		cfg.Env, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DBName)

	ctx := context.Background()

	db, err := database.New(ctx, &cfg.Postgres)

	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	}

	migrationsPath := filepath.Join(wd, "migrations")
	if err := db.RunMigrations(migrationsPath); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	log.Println("Successfully connected to database and ran migrations!")

	// TODO: Запустить gRPC сервер
	// TODO: Подключиться к RabbitMQ
}
