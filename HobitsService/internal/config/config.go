// internal/config/config.go
package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
)

type Config struct {
	Env      string `env:"ENV" env-default:"local"`
	GRPC     GRPCConfig
	Postgres PostgresConfig
	RabbitMQ RabbitMQConfig
}

type GRPCConfig struct {
	Port int `env:"GRPC_PORT" env-default:"50051"`
}

type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port     int    `env:"POSTGRES_PORT" env-default:"5432"`
	User     string `env:"POSTGRES_USER" env-default:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" env-default:"postgres"`
	DBName   string `env:"POSTGRES_DBNAME" env-default:"mydb"`
}

type RabbitMQConfig struct {
	URL string `env:"RABBITMQ_URL" env-default:"amqp://guest:guest@localhost:5672/"`
}

func MustLoad() *Config {
	var cfg Config

	// Только читаем ENV, без yaml
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read env: %s", err)
	}

	return &cfg
}
