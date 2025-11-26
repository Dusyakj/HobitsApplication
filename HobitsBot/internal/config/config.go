package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	GRPCServerAddr   string
	LogLevel         string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		GRPCServerAddr:   getEnv("GRPC_SERVER_ADDR", "localhost:50051"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
