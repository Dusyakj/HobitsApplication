# HobitsService - Описание Проекта

## Содержание
1. [Обзор](#обзор)
2. [Сущности БД](#сущности-бд)
3. [Domain Сущности](#domain-сущности)
4. [Сервисы](#сервисы)
5. [Архитектура Проекта](#архитектура-проекта)

---

## Обзор

**HobitsService** - это микросервис на Go для отслеживания привычек с интеграцией Telegram. Проект использует Clean Architecture, gRPC API и PostgreSQL.

**Ключевые возможности:**
- Управление привычками с гибким расписанием (ежедневно/еженедельно/ежемесячно)
- Интеллектуальное отслеживание серий выполнения (streaks)
- Автоматическая генерация напоминаний
- Система очереди для обработки пропущенных привычек
- Интеграция с Telegram
- Prometheus метрики и health checks

**Стек технологий:**
- Go 1.21+
- PostgreSQL + pgx/v5
- gRPC + Protocol Buffers
- Docker + docker-compose
- Zap logging
- Prometheus metrics

---

## Сущности БД

Проект использует PostgreSQL с 5 таблицами. Все миграции находятся в `migrations/`.

### 1. Users (Пользователи)

**Файл миграции:** `000001_init.up.sql`

**Описание:** Хранит информацию о пользователях Telegram.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | SERIAL PRIMARY KEY | Автоинкремент ID |
| `telegram_id` | BIGINT UNIQUE NOT NULL | ID пользователя в Telegram |
| `first_name` | VARCHAR(255) | Имя |
| `last_name` | VARCHAR(255) | Фамилия |
| `username` | VARCHAR(255) | Username в Telegram |
| `language_code` | VARCHAR(10) | Код языка (по умолчанию 'en') |
| `created_at` | TIMESTAMP | Дата создания |
| `updated_at` | TIMESTAMP | Дата обновления |

**Связи:**
- One-to-many: Habits, HabitLogs, HabitReminders, StreakResetQueue

---

### 2. Habits (Привычки)

**Файл миграции:** `000002_habits_tables.up.sql`

**Описание:** Основная таблица для хранения привычек пользователей.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | SERIAL PRIMARY KEY | ID привычки |
| `user_id` | INTEGER (FK → users) | Владелец привычки |
| `name` | VARCHAR(255) NOT NULL | Название привычки |
| `description` | TEXT | Описание (опционально) |
| `goal` | VARCHAR(255) | Цель (опционально) |
| `frequency` | VARCHAR(50) NOT NULL | Частота: 'daily', 'weekly', 'monthly' |
| `weekly_days` | VARCHAR(13) | Дни для weekly (формат: "1,3,5") |
| `monthly_days` | VARCHAR(255) | Дни для monthly (формат: "1,15,28") |
| `current_streak` | INTEGER (default 0) | Текущая серия выполнений |
| `best_streak` | INTEGER (default 0) | Лучшая серия |
| `last_completed_date` | DATE | Дата последнего выполнения |
| `last_checked_date` | DATE | Дата последней проверки |
| `is_active` | BOOLEAN (default TRUE) | Активна ли привычка |
| `is_completed` | BOOLEAN (default FALSE) | Достигнута ли цель |
| `created_at` | TIMESTAMP | Дата создания |
| `updated_at` | TIMESTAMP | Дата обновления |
| `completed_at` | TIMESTAMP | Дата достижения цели |

**Ограничения:**
- CHECK: frequency IN ('daily', 'weekly', 'monthly')
- FOREIGN KEY: user_id → users(id) ON DELETE CASCADE

**Индексы:**
- `idx_habits_user_id` на user_id
- `idx_habits_is_active` на is_active

**Связи:**
- Many-to-one: Users
- One-to-many: HabitLogs, HabitReminders, StreakResetQueue

---

### 3. Habit Logs (Логи выполнения)

**Файл миграции:** `000002_habits_tables.up.sql`

**Описание:** Записи о выполнении привычек.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | SERIAL PRIMARY KEY | ID записи |
| `habit_id` | INTEGER (FK → habits) | Связь с привычкой |
| `user_id` | INTEGER (FK → users) | Пользователь |
| `comment` | TEXT | Комментарий (опционально) |
| `logged_date` | DATE NOT NULL | Дата выполнения |
| `logged_at` | TIMESTAMP | Время создания записи |

**Ограничения:**
- UNIQUE: (habit_id, logged_date) - одна запись на привычку в день
- FOREIGN KEY: habit_id, user_id с CASCADE delete

**Индексы:**
- `idx_habit_logs_habit_id`
- `idx_habit_logs_user_id`
- `idx_habit_logs_logged_date`

**Связи:**
- Many-to-one: Habits, Users

---

### 4. Habit Reminders (Напоминания)

**Файл миграции:** `000002_habits_tables.up.sql`

**Описание:** Напоминания о привычках.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | SERIAL PRIMARY KEY | ID напоминания |
| `habit_id` | INTEGER (FK → habits) | Связь с привычкой |
| `user_id` | INTEGER (FK → users) | Пользователь |
| `reminder_date` | DATE NOT NULL | Дата напоминания |
| `is_completed` | BOOLEAN (default FALSE) | Выполнено ли |
| `sent_at` | TIMESTAMP | Время отправки |

**Ограничения:**
- UNIQUE: (habit_id, reminder_date) - одно напоминание на привычку в день
- FOREIGN KEY: habit_id, user_id с CASCADE delete

**Индексы:**
- `idx_habit_reminders_user_id`
- `idx_habit_reminders_reminder_date`

**Связи:**
- Many-to-one: Habits, Users

---

### 5. Streak Reset Queue (Очередь сброса серий)

**Файл миграции:** `000002_habits_tables.up.sql`

**Описание:** Очередь для обработки сброса серий при пропусках.

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | SERIAL PRIMARY KEY | ID записи в очереди |
| `habit_id` | INTEGER (FK → habits) | Привычка для сброса |
| `user_id` | INTEGER (FK → users) | Владелец |
| `reset_date` | DATE NOT NULL | Дата сброса |
| `processed` | BOOLEAN (default FALSE) | Обработано ли |
| `processed_at` | TIMESTAMP | Время обработки |
| `previous_streak` | INTEGER | Серия до сброса (для аудита) |
| `created_at` | TIMESTAMP | Время создания |

**Ограничения:**
- UNIQUE: (habit_id, reset_date) - один сброс на привычку в день
- FOREIGN KEY: habit_id, user_id с CASCADE delete

**Индексы:**
- `idx_streak_reset_queue_reset_date`
- `idx_streak_reset_queue_processed`

**Связи:**
- Many-to-one: Habits, Users

---

## Domain Сущности

Все domain сущности находятся в `internal/domain/`. Это богатые доменные модели с бизнес-логикой.

### 1. User (domain/user.go)

**Описание:** Представляет пользователя Telegram в системе.

**Поля:**
```go
type User struct {
    ID           int
    TelegramID   int64
    FirstName    string
    LastName     string
    Username     string
    LanguageCode string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**Конструктор:**
- `NewUser(telegramID, firstName, lastName, username, languageCode)` - Создание из данных Telegram

**Цель:** Управление учетными записями пользователей Telegram.

---

### 2. Habit (domain/habit.go)

**Описание:** Основная бизнес-сущность для привычек.

**Поля:**
```go
type Habit struct {
    ID                int
    UserID            int
    Name              string
    Description       sql.NullString
    Goal              sql.NullString
    Frequency         HabitFrequency // daily/weekly/monthly
    WeeklyDays        sql.NullString // "1,3,5"
    MonthlyDays       sql.NullString // "1,15,28"
    CurrentStreak     int
    BestStreak        int
    LastCompletedDate sql.NullTime
    LastCheckedDate   sql.NullTime
    IsActive          bool
    IsCompleted       bool
    CreatedAt         time.Time
    UpdatedAt         time.Time
    CompletedAt       sql.NullTime
}
```

**Константы:**
```go
const (
    FrequencyDaily   HabitFrequency = "daily"
    FrequencyWeekly  HabitFrequency = "weekly"
    FrequencyMonthly HabitFrequency = "monthly"
)
```

**Методы:**
- `NewHabit(userID, name, frequency)` - Конструктор
- `SetDescription(description)` - Установить описание
- `SetGoal(goal)` - Установить цель
- `SetWeeklyDays(days []int)` - Установить дни недели (формат: [1,3,5])
- `SetMonthlyDays(days []int)` - Установить дни месяца (формат: [1,15,28])
- `Deactivate()` - Деактивировать привычку
- `Activate()` - Активировать привычку
- `IncreaseStreak()` - Увеличить серию (обновляет BestStreak при необходимости)
- `ResetStreak()` - Сбросить серию до 0
- `MarkAsCompleted()` - Пометить цель как достигнутую
- `UpdateLastCompletedDate(date)` - Обновить дату последнего выполнения
- `UpdateLastCheckedDate(date)` - Обновить дату последней проверки

**Цель:** Управление привычками с логикой расписания и отслеживания серий.

---

### 3. HabitLog (domain/habit_log.go)

**Описание:** Запись о выполнении привычки.

**Поля:**
```go
type HabitLog struct {
    ID         int
    HabitID    int
    UserID     int
    Comment    sql.NullString
    LoggedDate time.Time
    LoggedAt   time.Time
}
```

**Методы:**
- `NewHabitLog(habitID, userID, loggedDate, comment)` - Конструктор
- `GetComment()` - Возвращает комментарий или пустую строку

**Цель:** Ежедневные записи выполнения с опциональными заметками.

---

### 4. HabitReminder (domain/habit_reminder.go)

**Описание:** Напоминание о привычке.

**Поля:**
```go
type HabitReminder struct {
    ID           int
    HabitID      int
    UserID       int
    ReminderDate sql.NullTime
    IsCompleted  bool
    SentAt       time.Time
}
```

**Методы:**
- `NewHabitReminder(habitID, userID, reminderDate)` - Конструктор
- `MarkAsCompleted()` - Пометить как выполненное
- `MarkAsIncomplete()` - Пометить как невыполненное

**Цель:** Отслеживание отправленных напоминаний.

---

### 5. StreakResetQueue (domain/streak_reset_queue.go)

**Описание:** Запись в очереди сброса серий.

**Поля:**
```go
type StreakResetQueue struct {
    ID             int
    HabitID        int
    UserID         int
    ResetDate      sql.NullTime
    Processed      bool
    ProcessedAt    sql.NullTime
    PreviousStreak sql.NullInt64
    CreatedAt      time.Time
}
```

**Методы:**
- `NewStreakResetQueue(habitID, userID, resetDate)` - Конструктор
- `MarkAsProcessed(previousStreak)` - Пометить как обработанное с аудитом
- `GetResetDate()` - Получить дату сброса
- `GetPreviousStreak()` - Получить предыдущую серию

**Цель:** Отложенная обработка сбросов с аудитом изменений.

---

## Сервисы

Все сервисы находятся в `internal/service/`. Это слой бизнес-логики.

### 1. UserService (service/user.go)

**Ответственность:** Управление учетными записями пользователей и интеграция с Telegram.

**Работает с:** User

**Зависимости:** UserRepository

**Ключевые методы:**

| Метод | Описание |
|-------|----------|
| `GetOrCreateUser(ctx, telegramID, firstName, lastName, username, languageCode)` | Находит существующего пользователя по Telegram ID или создает нового |
| `GetUser(ctx, id)` | Получает пользователя по внутреннему ID |
| `GetUserByTelegramID(ctx, telegramID)` | Получает пользователя по Telegram ID |
| `GetAllUsers(ctx)` | Получает всех пользователей (для scheduler) |
| `UpdateUser(ctx, id, firstName, lastName, username, languageCode)` | Обновляет информацию о пользователе |
| `DeleteUser(ctx, id)` | Удаляет пользователя |

**Цель:** Управление пользователями для интеграции с Telegram ботом.

---

### 2. HabitService (service/habit.go)

**Ответственность:** Управление привычками (CRUD) и логика расписания.

**Работает с:** Habit

**Зависимости:** HabitRepository, HabitLogRepository, HabitReminderRepository

**Ключевые методы:**

| Метод | Описание |
|-------|----------|
| `CreateHabit(ctx, userID, name, frequency)` | Создает новую привычку |
| `GetHabit(ctx, habitID)` | Получает привычку по ID |
| `GetUserHabits(ctx, userID)` | Получает все привычки пользователя |
| `GetActiveUserHabits(ctx, userID)` | Получает только активные привычки |
| `GetAllActiveHabits(ctx)` | Получает все активные привычки (для scheduler) |
| `UpdateHabit(ctx, habit)` | Обновляет привычку |
| `DeactivateHabit(ctx, habitID)` | Деактивирует привычку (soft delete) |
| `ActivateHabit(ctx, habitID)` | Активирует привычку |
| `SetWeeklyDays(ctx, habitID, days)` | Устанавливает дни недели (валидирует frequency) |
| `SetMonthlyDays(ctx, habitID, days)` | Устанавливает дни месяца (валидирует frequency) |
| `GetScheduledDaysForToday(ctx, habitID)` | Проверяет, запланирована ли привычка на сегодня |
| `GetScheduledDaysBetween(ctx, habitID, from, to)` | Возвращает все запланированные даты в диапазоне |

**Внутренняя логика:**
- `isHabitScheduledForDate(habit, date)` - Бизнес-логика проверки расписания
- `containsDay(daysStr, day)` - Парсинг запланированных дней
- `goWeekdayToInt(weekday)` - Конвертация дней недели (0=Sunday → 1=Monday, 7=Sunday ISO)
- `daysToString(days)` - Преобразование массива в строку

**Цель:** Центральное управление привычками со сложной логикой расписания для daily/weekly/monthly частот.

---

### 3. LogService (service/log.go)

**Ответственность:** Логирование выполнения привычек и расчет серий.

**Работает с:** HabitLog, Habit, HabitReminder, StreakResetQueue

**Зависимости:** HabitLogRepository, HabitRepository, HabitReminderRepository, StreakResetQueueRepository, HabitService

**Ключевые методы:**

| Метод | Описание |
|-------|----------|
| `LogCompletion(ctx, habitID, userID, comment)` | Логирует выполнение привычки на сегодня<br>• Валидирует владельца<br>• Предотвращает дубликаты<br>• Обновляет серию<br>• Помечает напоминание как выполненное<br>• Удаляет из очереди сброса |
| `GetHabitLogs(ctx, habitID)` | Получает все логи для привычки |
| `GetHabitLogsByDateRange(ctx, habitID, from, to)` | Получает логи за период |
| `GetCompletionRate(ctx, habitID, from, to)` | Рассчитывает процент выполнения |

**Внутренняя логика:**
- `updateStreak(ctx, habit)` - Расчет серий:
  - Обрабатывает первое выполнение
  - Обнаруживает разрыв серии
  - Обновляет лучшую серию
- `isStreakBroken(habit, lastDate, today)` - Проверяет пропущенные дни по расписанию

**Цель:** Отслеживание выполнения с интеллектуальным управлением сериями на основе расписания.

---

### 4. ReminderService (service/reminder.go)

**Ответственность:** Генерация и управление ежедневными напоминаниями.

**Работает с:** HabitReminder, Habit

**Зависимости:** HabitReminderRepository, HabitRepository, HabitService

**Ключевые методы:**

| Метод | Описание |
|-------|----------|
| `CreateReminder(ctx, habitID, userID, reminderDate)` | Создает одно напоминание |
| `GenerateRemindersForToday(ctx, userID)` | Генерирует все напоминания для запланированных на сегодня привычек<br>• Проверяет существующие напоминания<br>• Получает активные привычки<br>• Создает напоминания только для привычек, запланированных на сегодня<br>• Пропускает дубликаты |
| `GetRemindersByDate(ctx, date)` | Получает все напоминания на дату |
| `GetRemindersByUserAndDate(ctx, userID, date)` | Получает напоминания пользователя на дату |
| `MarkReminderAsCompleted(ctx, reminderID)` | Помечает напоминание как выполненное |
| `MarkReminderAsIncomplete(ctx, reminderID)` | Помечает напоминание как невыполненное |

**Цель:** Ежедневная генерация напоминаний на основе расписания привычек.

---

### 5. StreakResetService (service/streak_reset.go)

**Ответственность:** Управление очередью сброса серий для пропущенных привычек.

**Работает с:** StreakResetQueue, Habit, HabitLog, HabitReminder

**Зависимости:** StreakResetQueueRepository, HabitRepository, HabitLogRepository, HabitReminderRepository, HabitService

**Ключевые методы:**

| Метод | Описание |
|-------|----------|
| `CheckAndQueueStreakResets(ctx)` | Ежедневная задача для постановки сбросов в очередь (сейчас закомментировано) |
| `CheckHabitStreak(ctx, habitID)` | Проверяет одну привычку на пропущенные запланированные дни<br>• Получает запланированные дни с последней проверки<br>• Проверяет наличие лога для каждого дня<br>• Ставит сброс в очередь для пропущенных дней<br>• Обновляет last_checked_date |
| `ProcessQueueEntries(ctx)` | Обрабатывает все необработанные сбросы |
| `GetQueueEntry(ctx, entryID)` | Получает запись из очереди |
| `GetUnprocessedQueueEntries(ctx)` | Получает необработанные записи |

**Внутренняя логика:**
- `processQueueEntry(ctx, entry)` - Обработка одного сброса:
  - Сохраняет предыдущую серию для аудита
  - Обновляет лучшую серию при необходимости
  - Сбрасывает текущую серию до 0
  - Помечает напоминание как невыполненное

**Цель:** Система отложенного сброса серий с аудитом.

**Примечание:** Реализация scheduler сейчас закомментирована.

---

## Архитектура Проекта

### Структура каталогов

```
HobitsService/
├── cmd/
│   └── server/
│       └── main.go                 # Точка входа приложения
├── internal/
│   ├── app/
│   │   └── app.go                  # DI-контейнер
│   ├── config/
│   │   └── config.go               # Управление конфигурацией
│   ├── delivery/
│   │   ├── grpc/                   # gRPC handlers (уровень представления)
│   │   │   ├── server.go
│   │   │   ├── converters.go       # Конвертация Proto ↔ Domain
│   │   │   ├── habit_handler.go
│   │   │   ├── log_handler.go
│   │   │   ├── reminder_handler.go
│   │   │   └── user_handler.go
│   │   └── rabbitmq/               # Интеграция с очередями (в будущем)
│   ├── domain/                     # Бизнес-сущности (ядро)
│   │   ├── user.go
│   │   ├── habit.go
│   │   ├── habit_log.go
│   │   ├── habit_reminder.go
│   │   └── streak_reset_queue.go
│   ├── infrastructure/             # Внешние зависимости
│   │   ├── database/
│   │   │   └── postgres.go         # Подключение к БД и миграции
│   │   ├── scheduler/
│   │   │   └── scheduler.go        # Scheduled tasks (закомментировано)
│   │   └── queue/                  # Очереди сообщений (в будущем)
│   ├── logger/
│   │   └── logger.go               # Инфраструктура логирования
│   ├── repository/                 # Уровень доступа к данным
│   │   ├── repository.go           # Интерфейсы репозиториев
│   │   └── postgres/               # PostgreSQL реализации
│   │       ├── user.go
│   │       ├── habit.go
│   │       ├── habit_log.go
│   │       ├── habit_reminder.go
│   │       └── streak_reset_queue.go
│   └── service/                    # Уровень бизнес-логики
│       ├── user.go
│       ├── habit.go
│       ├── log.go
│       ├── reminder.go
│       └── streak_reset.go
├── migrations/                     # Миграции БД
│   ├── 000001_init.up.sql
│   ├── 000001_init.down.sql
│   ├── 000002_habits_tables.up.sql
│   └── 000002_habits_tables.down.sql
├── gen/                            # Сгенерированный protobuf код
│   └── go/HobitsService/gen/go/hobbits/api/v1/
│       ├── common.pb.go
│       ├── habit_service.pb.go
│       ├── habit_service_grpc.pb.go
│       ├── log_service.pb.go
│       ├── log_service_grpc.pb.go
│       ├── reminder_service.pb.go
│       ├── reminder_service_grpc.pb.go
│       ├── user_service.pb.go
│       └── user_service_grpc.pb.go
├── proto/                          # Protobuf определения
├── logs/                           # Логи приложения
├── pkg/                            # Общие пакеты
├── config.yaml                     # Файл конфигурации
├── docker-compose.yml              # Docker services setup
├── Dockerfile                      # Service container
├── generate.sh                     # Скрипт генерации protobuf
├── go.mod                          # Go зависимости
└── prometheus.yml                  # Конфигурация метрик
```

### Архитектурные слои

#### 1. Уровень представления (Presentation Layer)
**Путь:** `internal/delivery/grpc/`

- gRPC handlers, реализующие protobuf service definitions
- Конвертация между protobuf сообщениями и domain сущностями
- HTTP/2 коммуникация с клиентами
- 4 сервиса: User, Habit, Log, Reminder

#### 2. Уровень бизнес-логики (Business Logic Layer)
**Путь:** `internal/service/`

- Чистая бизнес-логика без инфраструктурных зависимостей
- Оркестрация операций репозиториев
- Реализация сложных доменных правил (расчет серий, расписание)
- Нет прямого доступа к базе данных

#### 3. Уровень доступа к данным (Data Access Layer)
**Путь:** `internal/repository/`

- Repository pattern с интерфейсами
- PostgreSQL реализации с использованием pgx драйвера
- Абстракция над операциями с БД
- Легкое тестирование с моками

#### 4. Доменный слой (Domain Layer)
**Путь:** `internal/domain/`

- Основные бизнес-сущности
- Доменные методы и валидация
- Нет зависимостей от других слоев
- Богатая доменная модель с поведением

#### 5. Инфраструктурный слой (Infrastructure Layer)
**Путь:** `internal/infrastructure/`

- Управление подключением к БД
- Task scheduling (сейчас отключен)
- Интеграции с внешними сервисами
- Метрики и мониторинг

### Используемые паттерны проектирования

1. **Repository Pattern** - Абстракция доступа к данным
2. **Dependency Injection** - Через `app.go` контейнер
3. **Clean Architecture** - Правило зависимостей (внутренние слои не знают о внешних)
4. **Service Layer Pattern** - Разделение бизнес-логики
5. **Factory Pattern** - Конструкторы (NewX)
6. **Strategy Pattern** - Разное поведение для разных частот (daily/weekly/monthly)

### Жизненный цикл приложения

1. Загрузка конфигурации из environment
2. Инициализация logger (zap)
3. Инициализация метрик (Prometheus)
4. Подключение к PostgreSQL
5. Запуск миграций
6. Инициализация репозиториев
7. Инициализация сервисов
8. Запуск HTTP сервера (порт 8080) для /metrics и /health
9. Запуск gRPC сервера (порт 50051)
10. Запуск scheduler (сейчас отключен)
11. Ожидание сигнала завершения
12. Graceful shutdown с таймаутом 30 секунд

### Ключевые особенности

✅ **Интеграция с Telegram** - Управление пользователями через Telegram IDs
✅ **Гибкое расписание** - Daily, weekly (конкретные дни), monthly (конкретные даты)
✅ **Отслеживание серий** - Текущая и лучшая серия с интеллектуальным расчетом
✅ **Система напоминаний** - Автогенерация ежедневных напоминаний по расписанию
✅ **Очередь сбросов** - Отложенная обработка сбросов серий с аудитом
✅ **Метрики** - Интеграция с Prometheus для мониторинга
✅ **Health Checks** - HTTP сервер для /health и /metrics
✅ **Graceful Shutdown** - Обработка сигналов с корректным завершением

### Дизайн схемы БД

- **Нормализация:** 3NF, минимальная избыточность
- **Референциальная целостность:** Foreign keys с CASCADE delete
- **Ограничения:** CHECK constraints для enum'ов, UNIQUE для бизнес-правил
- **Индексирование:** Стратегические индексы на внешних ключах и колонках запросов
- **Аудит:** Timestamps на всех таблицах, отслеживание previous_streak

---

## Резюме

**HobitsService** - хорошо спроектированный Go микросервис для отслеживания привычек со следующими характеристиками:

**Статистика:**
- 5 таблиц БД: users, habits, habit_logs, habit_reminders, streak_reset_queue
- 5 Domain сущностей: богатые доменные модели с бизнес-логикой
- 5 классов сервисов: UserService, HabitService, LogService, ReminderService, StreakResetService
- 5 реализаций репозиториев: PostgreSQL уровень доступа к данным
- 4 gRPC сервиса: User, Habit, Log, Reminder APIs

**Характеристики:**
- ✅ Clean Architecture с четким разделением ответственности
- ✅ Production-ready: метрики, логирование, health checks, graceful shutdown
- ✅ Гибкое расписание: daily, weekly, monthly частоты
- ✅ Интеллектуальные серии: автоматический расчет на основе соблюдения расписания
- ✅ Обработка на основе очередей: отложенные сбросы с аудитом

Кодовая база чистая, хорошо организованная, следует лучшим практикам Go с dependency injection, дизайном на основе интерфейсов и комплексной обработкой ошибок.

---

**Для конвертации в PDF:**
- Pandoc: `pandoc PROJECT_OVERVIEW.md -o PROJECT_OVERVIEW.pdf`
- VS Code: установите расширение "Markdown PDF"
- Онлайн: https://www.markdowntopdf.com/
