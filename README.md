# 🎯 HobitsApplication - Полнофункциональная система отслеживания привычек

Комплексное приложение для отслеживания ежедневных привычек с Telegram ботом, backend сервисом и веб-интерфейсом.

## 📋 Архитектура

```
┌─────────────────────────────────┐
│   Telegram Users                │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│   HobitsBot (Telegram)          │
│   - Message handling            │
│   - Inline keyboards            │
│   - State management            │
└────────────┬────────────────────┘
             │ gRPC
             ▼
┌─────────────────────────────────┐
│   HobitsService (Backend)       │
│   - Habit management            │
│   - User management             │
│   - Logging & analytics         │
│   - gRPC API                    │
└────────────┬────────────────────┘
             │
    ┌────────┼────────┐
    │        │        │
    ▼        ▼        ▼
┌────────┬────────┬────────┐
│  Postgres  RabbitMQ      │
│  (Database)(Queue)       │
└────────┬────────┬────────┘
    │        │
    ▼        ▼
┌──────────────────────┐
│  Prometheus & Grafana│
│  (Monitoring)        │
└──────────────────────┘
```

## 🚀 Быстрый старт

### Предварительные требования
- Docker 20.10+
- Docker Compose 2.0+
- Git

### 1. Клонирование репозитория

```bash
git clone <repository-url>
cd HobitsApplication
```

### 2. Конфигурация

Скопируйте пример файла конфигурации:

```bash
cp .env.example .env
```

Отредактируйте `.env` файл и установите свои значения:

```env
# Обязательно установить Telegram Bot Token
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here

# Остальные параметры имеют значения по умолчанию
DB_USER=hobits
DB_PASSWORD=password
DB_NAME=hobits
```

### 3. Запуск всех сервисов

```bash
docker-compose up -d
```

Это запустит:
- **PostgreSQL** (5432) - база данных
- **RabbitMQ** (5672) - очередь сообщений
- **HobitsService** (50051, 8080) - backend gRPC сервер
- **HobitsBot** - Telegram бот
- **Prometheus** (9090) - сбор метрик
- **Grafana** (3000) - визуализация метрик

### 4. Проверка статуса

```bash
docker-compose ps
```

Все контейнеры должны быть в статусе `running`.

## 📱 Использование Telegram Бота

1. Найдите вашего бота в Telegram
2. Напишите `/start` для инициализации
3. Используйте команды:
   - `/add` - добавить новую привычку
   - `/habits` - список всех привычек
   - `/today` - напоминания на сегодня
   - `/stats` - статистика привычек
   - `/help` - справка

Подробнее смотрите в [HobitsBot/EXAMPLES.md](HobitsBot/EXAMPLES.md)

## 📊 Мониторинг

### Grafana Dashboard
- **URL**: http://localhost:3000
- **Логин**: admin
- **Пароль**: admin (из `.env`)

### Prometheus Metrics
- **URL**: http://localhost:9090

### RabbitMQ Management
- **URL**: http://localhost:15672
- **Логин**: guest
- **Пароль**: guest

## 📁 Структура проекта

```
HobitsApplication/
├── HobitsService/          # Backend сервер
│   ├── cmd/                # Точка входа
│   ├── internal/           # Внутренний код
│   ├── migrations/         # Миграции БД
│   ├── Dockerfile
│   └── docker-compose.yml
│
├── HobitsBot/              # Telegram бот
│   ├── cmd/                # Точка входа
│   ├── internal/
│   │   ├── bot/            # Логика бота
│   │   ├── service/        # gRPC клиенты
│   │   ├── config/         # Конфигурация
│   │   └── logger/         # Логирование
│   ├── Dockerfile
│   └── docker-compose.yml
│
├── HobitsRestController/   # REST контроллер (если требуется)
│
├── docker-compose.yml      # Главный compose файл
├── .env.example            # Пример конфигурации
└── README.md               # Этот файл
```

## 🛠️ Разработка

### Запуск одного сервиса локально

#### HobitsService
```bash
cd HobitsService
make install-deps
make run
```

#### HobitsBot
```bash
cd HobitsBot
make install-deps
cp .env.example .env
# Отредактируйте .env с нужными значениями
make run
```

### Команды Make

#### HobitsService
```bash
make help          # Показать все команды
make build         # Собрать приложение
make run           # Запустить локально
make test          # Запустить тесты
make docker-build  # Собрать Docker образ
```

#### HobitsBot
```bash
make help          # Показать все команды
make build         # Собрать бот
make run           # Запустить локально
make docker-build  # Собрать Docker образ
make docker-run    # Запустить в Docker
```

## 📚 Документация

- [HobitsBot Architecture](HobitsBot/ARCHITECTURE.md) - Архитектура бота
- [HobitsBot Examples](HobitsBot/EXAMPLES.md) - Примеры использования
- [HobitsBot README](HobitsBot/README.md) - Документация бота

## 🔧 Управление Docker

### Просмотр логов
```bash
# Все сервисы
docker-compose logs -f

# Конкретный сервис
docker-compose logs -f hobitsbot
docker-compose logs -f hobits-service
docker-compose logs -f postgres
```

### Остановка сервисов
```bash
docker-compose down
```

### Полная очистка (включая данные)
```bash
docker-compose down -v
```

### Пересборка образов
```bash
docker-compose build --no-cache
docker-compose up -d
```

## 🗄️ База данных

### Доступ к PostgreSQL
```bash
docker-compose exec postgres psql -U hobits -d hobits
```

### Просмотр логов миграций
```bash
docker-compose logs hobits-service | grep -i migration
```

## 🐛 Решение проблем

### Бот не подключается к сервису
1. Проверьте, запущен ли HobitsService: `docker-compose ps`
2. Проверьте логи: `docker-compose logs hobits-service`
3. Убедитесь, что GRPC_SERVER_ADDR установлен правильно в `.env`

### Ошибка подключения к БД
1. Проверьте статус PostgreSQL: `docker-compose ps postgres`
2. Убедитесь, что переменные БД совпадают в `.env`
3. Проверьте логи: `docker-compose logs postgres`

### Telegram бот не реагирует
1. Проверьте TELEGRAM_BOT_TOKEN в `.env`
2. Убедитесь, что токен скопирован правильно без пробелов
3. Проверьте логи: `docker-compose logs hobitsbot`

## 📈 Производительность

### Оптимизация для больших объемов

- **Кеширование**: Prometheus хранит метрики 30 дней
- **Масштабирование**: Запускайте несколько реплик бота через Docker Swarm/Kubernetes
- **Оптимизация БД**: Используйте индексы на часто запрашиваемых полях

### Мониторинг ресурсов
```bash
docker stats
```

## 🔐 Безопасность

⚠️ **Важно**: Никогда не коммитьте реальные значения в `.env`:
- TELEGRAM_BOT_TOKEN
- DB_PASSWORD
- GRAFANA_PASSWORD

Используйте `.env.example` как шаблон и хранитесь в `.gitignore`.

## 📝 Лицензия

[Укажите лицензию проекта]

## 👥 Автор

[Укажите информацию об авторе]

---

**Готово к использованию!** 🚀 Запустите `docker-compose up -d` и начните отслеживать свои привычки!
