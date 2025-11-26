# 📁 Структура проекта HobitsApplication

```
HobitsApplication/
│
├── 📄 README.md                    # Главная документация
├── 📄 QUICKSTART.md               # Быстрый старт (5 минут)
├── 📄 DEPLOYMENT.md               # Развертывание (dev, prod, k8s)
├── 📄 PROJECT_STRUCTURE.md        # Этот файл
│
├── 🔧 docker-compose.yml          # Главный Docker Compose (запуск всех сервисов)
├── 🔧 Makefile                    # Команды для управления проектом
├── 🔧 .env.example                # Пример конфигурации
├── 🔧 .gitignore                  # Git ignore правила
│
├── 📦 HobitsService/              # Backend сервис
│   ├── cmd/
│   │   └── server/
│   │       └── main.go            # Точка входа сервера
│   ├── internal/
│   │   ├── app/                   # Основное приложение
│   │   ├── domain/                # Бизнес логика (модели, интерфейсы)
│   │   ├── delivery/              # HTTP и gRPC хендлеры
│   │   ├── service/               # Сервис слой
│   │   ├── repository/            # Работа с БД
│   │   ├── infrastructure/        # Внешние зависимости
│   │   ├── logger/                # Логирование
│   │   └── metrics/               # Prometheus метрики
│   ├── migrations/                # Миграции БД
│   ├── proto/                     # Protobuf определения
│   ├── gen/                       # Сгенерированный код из proto
│   ├── grafana/                   # Grafana конфигурация
│   ├── go.mod / go.sum            # Go зависимости
│   ├── Dockerfile                 # Docker образ для сервиса
│   ├── docker-compose.yml         # Старый compose (для локальной разработки)
│   ├── Makefile                   # Команды для разработки
│   └── README.md                  # Документация сервиса
│
├── 🤖 HobitsBot/                  # Telegram бот
│   ├── cmd/
│   │   └── main.go                # Точка входа бота
│   ├── internal/
│   │   ├── bot/
│   │   │   ├── bot.go             # Основная логика бота
│   │   │   ├── handlers.go        # Обработчики команд
│   │   │   ├── keyboards.go       # Inline клавиатуры
│   │   │   ├── formatter.go       # Форматирование сообщений
│   │   │   └── context.go         # Управление состоянием
│   │   ├── service/               # gRPC клиенты
│   │   │   ├── habit_service.go
│   │   │   ├── user_service.go
│   │   │   ├── log_service.go
│   │   │   └── reminder_service.go
│   │   ├── config/                # Конфигурация
│   │   ├── logger/                # Логирование
│   │   └── grpc/                  # gRPC клиент
│   ├── go.mod / go.sum            # Go зависимости
│   ├── Dockerfile                 # Docker образ для бота
│   ├── docker-compose.yml         # Старый compose (для локальной разработки)
│   ├── Makefile                   # Команды для разработки
│   ├── .env.example               # Пример конфигурации
│   ├── ARCHITECTURE.md            # Архитектура бота
│   ├── EXAMPLES.md                # Примеры использования
│   └── README.md                  # Документация бота
│
└── 🌐 HobitsRestController/       # REST контроллер (опционально)
    └── ...
```

## 📋 Назначение файлов в корне

| Файл | Назначение |
|------|-----------|
| `README.md` | Главная документация со всей информацией |
| `QUICKSTART.md` | Быстрый старт за 5 минут |
| `DEPLOYMENT.md` | Подробное руководство по развертыванию |
| `docker-compose.yml` | Главный файл для запуска всех сервисов |
| `Makefile` | Удобные команды для управления проектом |
| `.env.example` | Шаблон переменных окружения |
| `.gitignore` | Правила для git (исключение файлов) |

## 🗂️ Структура HobitsService

### Слои архитектуры (Clean Architecture)

```
HobitsService/
├── cmd/           # Точки входа (main functions)
├── internal/
│   ├── domain/    # Бизнес логика (независимо от фреймворков)
│   ├── service/   # Логика приложения
│   ├── delivery/  # HTTP/gRPC хендлеры (внешний интерфейс)
│   └── repository/# Работа с БД (внутренний интерфейс)
```

### Ключевые компоненты

- **Protobuf API** (`proto/`) - определение gRPC контрактов
- **PostgreSQL** - основная база данных
- **RabbitMQ** - очередь сообщений для асинхронных операций
- **Prometheus/Grafana** - мониторинг и метрики

## 🤖 Структура HobitsBot

### Слои архитектуры

```
HobitsBot/
├── cmd/           # Точка входа приложения
├── internal/
│   ├── bot/       # Логика обработки обновлений от Telegram
│   ├── service/   # gRPC клиенты для общения с backend
│   ├── config/    # Загрузка конфигурации
│   ├── logger/    # Логирование
│   └── grpc/      # gRPC клиент обертка
```

### Ключевые компоненты

- **Telegram Bot API** - интеграция с Telegram
- **gRPC Clients** - коммуникация с HobitsService
- **State Management** - отслеживание состояния пользователя
- **Inline Keyboards** - интерактивные кнопки

## 🔄 Поток данных

```
Telegram User
    ↓
TelegramAPI
    ↓
HobitsBot (handlers.go)
    ↓
gRPC Clients (service/)
    ↓
HobitsService (delivery/)
    ↓
HobitsService (service/)
    ↓
PostgreSQL Database
```

## 🚀 Как использовать структуру

### Для быстрого старта
1. Читать `QUICKSTART.md`
2. Запустить `docker-compose up -d`
3. Открыть бота в Telegram

### Для разработки
1. Читать `README.md`
2. Перейти в папку сервиса (`HobitsBot` или `HobitsService`)
3. Использовать `make` команды
4. Изучить документацию в папке

### Для production
1. Читать `DEPLOYMENT.md`
2. Выбрать тип развертывания (Docker, Kubernetes и т.д.)
3. Следовать инструкциям по безопасности

## 📚 Документация

### В корне проекта
- `README.md` - полная документация
- `QUICKSTART.md` - быстрый старт
- `DEPLOYMENT.md` - продакшн развертывание

### В папке HobitsService
- `README.md` - документация сервиса
- (внутри `internal/` есть комментарии в коде)

### В папке HobitsBot
- `README.md` - документация бота
- `ARCHITECTURE.md` - архитектура и дизайн
- `EXAMPLES.md` - примеры использования

## 🔧 Команды

### В корне проекта
```bash
make help         # Показать все команды
make up           # Запустить все сервисы
make down         # Остановить все сервисы
make logs         # Просмотреть логи
```

### В папке HobitsService
```bash
make build        # Собрать приложение
make run          # Запустить локально
make test         # Запустить тесты
make docker-build # Собрать Docker образ
```

### В папке HobitsBot
```bash
make build        # Собрать бот
make run          # Запустить локально
make docker-build # Собрать Docker образ
```

## 📦 Docker образы

При запуске `docker-compose up -d` создаются образы:

| Образ | Источник | Контейнер |
|-------|----------|-----------|
| postgres:15-alpine | Docker Hub | hobits-postgres |
| rabbitmq:3 | Docker Hub | hobits-rabbitmq |
| custom hobits-service | ./HobitsService | hobits-service |
| custom hobitsbot | ./HobitsBot | hobitsbot |
| prom/prometheus | Docker Hub | hobits-prometheus |
| grafana/grafana | Docker Hub | hobits-grafana |

## 🗄️ БД структура

PostgreSQL содержит таблицы для:
- Users (пользователи)
- Habits (привычки)
- HabitLogs (логирование выполнений)
- Reminders (напоминания)

Миграции находятся в `HobitsService/migrations/`

## 🔐 Конфиденциальность

### Файлы, которые НЕ должны быть в git:
- `.env` (только `.env.example`)
- Ключи и сертификаты
- Пароли и токены
- Логи и dump файлы

### Проверяйте `.gitignore` перед коммитом!

---

**Используйте эту структуру как справочник для навигации по проекту.** 🚀
