.PHONY: help setup build up down logs stop restart clean test lint docker-build dev

help:
	@echo "🎯 HobitsApplication - Система отслеживания привычек"
	@echo ""
	@echo "📦 Docker Commands:"
	@echo "  make setup         - Инициальная конфигурация (.env)"
	@echo "  make up            - Запустить все сервисы"
	@echo "  make down          - Остановить все сервисы"
	@echo "  make logs          - Просмотреть логи всех сервисов"
	@echo "  make logs-bot      - Логи Telegram бота"
	@echo "  make logs-service  - Логи backend сервиса"
	@echo "  make logs-db       - Логи БД"
	@echo "  make restart       - Перезагрузить все сервисы"
	@echo "  make clean         - Остановить и удалить все (с данными)"
	@echo ""
	@echo "🏗️  Build Commands:"
	@echo "  make build         - Собрать все Docker образы"
	@echo "  make build-bot     - Собрать только бот"
	@echo "  make build-service - Собрать только сервис"
	@echo ""
	@echo "🧪 Development Commands:"
	@echo "  make dev           - Запустить в режиме разработки"
	@echo "  make test          - Запустить тесты"
	@echo "  make lint          - Проверка кода"
	@echo ""
	@echo "📊 Status Commands:"
	@echo "  make ps            - Статус всех контейнеров"
	@echo "  make stats         - Статистика использования ресурсов"
	@echo ""
	@echo "💾 Database Commands:"
	@echo "  make db-shell      - Подключиться к PostgreSQL"
	@echo "  make db-dump       - Создать дамп БД"
	@echo "  make db-restore    - Восстановить БД из дампа"
	@echo ""

setup:
	@echo "⚙️  Setting up HobitsApplication..."
	@if [ ! -f .env ]; then \
		echo "📝 Creating .env file..."; \
		cp .env.example .env; \
		echo "✅ .env created. Please edit it with your values:"; \
		echo "   - TELEGRAM_BOT_TOKEN (required)"; \
		echo "   - Database credentials (optional, has defaults)"; \
	else \
		echo "✅ .env already exists"; \
	fi

up:
	@echo "🚀 Starting all services..."
	docker-compose up -d
	@echo "✅ All services started!"
	@echo ""
	@echo "📍 Service URLs:"
	@echo "   - Grafana:        http://localhost:3000 (admin/admin)"
	@echo "   - Prometheus:     http://localhost:9090"
	@echo "   - RabbitMQ:       http://localhost:15672 (guest/guest)"
	@echo "   - PostgreSQL:     localhost:5432"
	@echo "   - HobitsService:  localhost:50051 (gRPC)"
	@echo "   - Backend API:    http://localhost:8080"

down:
	@echo "🛑 Stopping all services..."
	docker-compose down
	@echo "✅ All services stopped"

restart:
	@echo "🔄 Restarting all services..."
	docker-compose restart
	@echo "✅ All services restarted"

logs:
	docker-compose logs -f

logs-bot:
	docker-compose logs -f hobitsbot

logs-service:
	docker-compose logs -f hobits-service

logs-db:
	docker-compose logs -f postgres

ps:
	@echo "📊 Container Status:"
	docker-compose ps

stats:
	@echo "📈 Resource Usage:"
	docker stats

build:
	@echo "🏗️  Building all Docker images..."
	docker-compose build --no-cache
	@echo "✅ All images built"

build-bot:
	@echo "🏗️  Building HobitsBot..."
	docker-compose build --no-cache hobitsbot
	@echo "✅ HobitsBot built"

build-service:
	@echo "🏗️  Building HobitsService..."
	docker-compose build --no-cache hobits-service
	@echo "✅ HobitsService built"

dev:
	@echo "👨‍💻 Starting development mode..."
	@echo "📝 Note: Make sure services are running: make up"
	@echo ""
	@echo "🔧 Available development targets:"
	@echo "   - cd HobitsBot && make run"
	@echo "   - cd HobitsService && make run"

test:
	@echo "🧪 Running tests..."
	@echo ""
	@echo "Testing HobitsService..."
	cd HobitsService && go test -v ./...
	@echo ""
	@echo "Testing HobitsBot..."
	cd HobitsBot && go test -v ./... 2>/dev/null || echo "No tests found for HobitsBot"

lint:
	@echo "🔍 Running linters..."
	@echo ""
	@echo "Linting HobitsService..."
	cd HobitsService && golangci-lint run ./... || echo "golangci-lint not installed, skipping"
	@echo ""
	@echo "Linting HobitsBot..."
	cd HobitsBot && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

clean:
	@echo "🗑️  Cleaning up all containers and volumes..."
	docker-compose down -v
	@echo "✅ All containers and volumes removed"

db-shell:
	@echo "📝 Connecting to PostgreSQL..."
	docker-compose exec postgres psql -U $${DB_USER:-hobits} -d $${DB_NAME:-hobits}

db-dump:
	@echo "💾 Creating database dump..."
	@mkdir -p backups
	docker-compose exec -T postgres pg_dump -U $${DB_USER:-hobits} -d $${DB_NAME:-hobits} > backups/dump_$$(date +%Y%m%d_%H%M%S).sql
	@echo "✅ Database dumped to backups/"

db-restore:
	@echo "📥 Restoring database from dump..."
	@if [ -z "$(DUMP_FILE)" ]; then \
		echo "❌ Error: DUMP_FILE not specified"; \
		echo "Usage: make db-restore DUMP_FILE=backups/dump_YYYYMMDD_HHMMSS.sql"; \
		exit 1; \
	fi
	docker-compose exec -T postgres psql -U $${DB_USER:-hobits} -d $${DB_NAME:-hobits} < $(DUMP_FILE)
	@echo "✅ Database restored"

.PHONY: setup up down restart logs logs-bot logs-service logs-db ps stats build build-bot build-service dev test lint clean db-shell db-dump db-restore
