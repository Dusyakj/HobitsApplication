.PHONY: help build run test clean install-deps fmt lint docker-build docker-run

help:
	@echo "HobitsBot - Telegram Bot для отслеживания привычек"
	@echo ""
	@echo "Доступные команды:"
	@echo "  make install-deps   - Установить зависимости"
	@echo "  make build          - Собрать приложение"
	@echo "  make run            - Запустить бот локально"
	@echo "  make test           - Запустить тесты"
	@echo "  make fmt            - Форматировать код"
	@echo "  make lint           - Проверить код"
	@echo "  make clean          - Очистить построенные файлы"
	@echo "  make docker-build   - Собрать Docker образ"
	@echo "  make docker-run     - Запустить в Docker"

install-deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

build:
	@echo "Building HobitsBot..."
	go build -o hobitsbot cmd/main.go

run: build
	@echo "Running HobitsBot..."
	@if [ ! -f .env ]; then \
		echo "Creating .env file from .env.example..."; \
		cp .env.example .env; \
		echo "Please edit .env with your configuration"; \
		exit 1; \
	fi
	./hobitsbot

test:
	@echo "Running tests..."
	go test -v ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Checking code with golangci-lint..."
	golangci-lint run ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -f hobitsbot
	go clean

docker-build:
	@echo "Building Docker image..."
	docker build -t hobitsbot:latest .

docker-run:
	@echo "Running HobitsBot in Docker..."
	docker run --env-file .env \
		--network host \
		hobitsbot:latest

docker-logs:
	@echo "Showing Docker logs..."
	docker logs -f $$(docker ps -q -f ancestor=hobitsbot:latest)

gen-mocks:
	@echo "Generating mocks..."
	mockgen -source=internal/service/habit_service.go -destination=internal/service/mocks/mock_habit.go -package=mocks

.PHONY: setup
setup: install-deps
	@echo "Setting up project..."
	@if [ ! -f .env ]; then cp .env.example .env; fi
	@echo "✅ Project is ready! Please configure .env and run 'make run'"
