package database

import (
	"HobitsService/internal/config"
	"HobitsService/internal/logger"
	"context"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"path/filepath"
	"runtime"
)

type Database struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, cfg *config.PostgresConfig) (*Database, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	logger.Debug(dsn)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &Database{Pool: pool}, nil
}

func (db *Database) RunMigrations(migrationsPath string) error {
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}

	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", absPath)
	}

	// Создаем правильный file:// URL для всех платформ
	var sourceURL string
	if runtime.GOOS == "windows" {
		// На Windows: C:\path -> file:///C:/path
		sourceURL = "file:///" + filepath.ToSlash(absPath)
	} else {
		// На Unix/Linux: /path -> file:///path
		sourceURL = "file://" + absPath
	}

	config := db.Pool.Config()
	dsn := config.ConnString()

	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (db *Database) Close() {
	db.Pool.Close()
}
