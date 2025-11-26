# 🏗️ Архитектура HobitsBot

## Общая архитектура

```
┌─────────────────────────────────────────────────────────┐
│                   Telegram Users                        │
└────────────────────────┬────────────────────────────────┘
                         │
                    (Telegram API)
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│                     HobitsBot                           │
│  ┌─────────────────────────────────────────────────┐   │
│  │  Update Handler Loop (Goroutines)              │   │
│  │  - Message Handler                            │   │
│  │  - Callback Query Handler                     │   │
│  │  - Error Recovery                             │   │
│  └─────────────────────────────────────────────────┘   │
│                         │                               │
│  ┌──────────────────────┴──────────────────────────┐   │
│  │                                                  │   │
│  ▼                                                  ▼   │
│  ┌──────────────────────┐          ┌──────────────┐   │
│  │  BotHandlers         │          │ ContextMgr   │   │
│  │  - Commands          │          │ - States     │   │
│  │  - Callbacks         │          │ - User Data  │   │
│  │  - Messages          │          │ - TTL Cache  │   │
│  └──────────────────────┘          └──────────────┘   │
│           │                              ▲              │
│           │        ┌────────────────────┘              │
│           │        │                                    │
│           ▼        ▼                                    │
│  ┌──────────────────────────────┐                     │
│  │   Formatters & Keyboards     │                     │
│  │ - Format Messages            │                     │
│  │ - Create Keyboards           │                     │
│  │ - Handle Buttons             │                     │
│  └──────────────────────────────┘                     │
│           │                                             │
└───────────┼─────────────────────────────────────────────┘
            │
      (gRPC API)
            │
            ▼
┌─────────────────────────────────────────────────────────┐
│          Services (gRPC Clients)                        │
│  ┌──────────────┐ ┌────────────┐ ┌────────────┐       │
│  │ HabitService │ │ UserService│ │ LogService │       │
│  └──────────────┘ └────────────┘ └────────────┘       │
│  ┌────────────────┐                                   │
│  │ReminderService │                                   │
│  └────────────────┘                                   │
└────────────────┬────────────────────────────────────────┘
                 │
           (gRPC Protocol)
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│              HobitsService (Backend)                    │
│  - Habit Management                                     │
│  - User Management                                      │
│  - Logging & Reminders                                  │
│  - Database Integration                                 │
└────────────────┬────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│                   Database                              │
│              (PostgreSQL)                               │
└─────────────────────────────────────────────────────────┘
```

## Компоненты

### 1. Bot Layer (`internal/bot`)

#### `bot.go` - Основной бот
- Инициализация Telegram API клиента
- Управление gRPC соединением
- Обработка обновлений
- Отправка сообщений

```go
type Bot struct {
    api    *tgbotapi.BotAPI       // Telegram API клиент
    grpc   *grpc.ClientConn       // gRPC соединение
    logger Logger                 // Логгер
}
```

#### `keyboards.go` - Клавиатуры и кнопки
- Создание inline кнопок
- Создание reply кнопок
- Динамическое формирование интерфейсов

```go
MainMenuKeyboard()           // Главное меню
HabitsListKeyboard()         // Список привычек
HabitActionsKeyboard()       // Действия с привычкой
FrequencyKeyboard()          // Выбор частоты
WeekdaysKeyboard()           // Выбор дней недели
DayNumbersKeyboard()         // Выбор дней месяца
ReminderTimeKeyboard()       // Выбор времени напоминания
```

#### `formatter.go` - Форматирование сообщений
- Форматирование приветствия
- Форматирование списков привычек
- Форматирование статистики
- Форматирование ошибок
- Использование эмодзи и Markdown

```go
FormatWelcomeMessage()
FormatHabitList()
FormatHabitDetail()
FormatStats()
FormatHabitCompleted()
```

#### `context.go` - Управление состоянием пользователя
- Сохранение состояния пользователя
- Временное хранилище данных
- Автоматическая очистка (TTL)
- Thread-safe операции (sync.RWMutex)

```go
type ContextManager struct {
    contexts map[int64]*UserContext    // Состояние пользователей
    mu       sync.RWMutex               // Синхронизация доступа
    ttl      time.Duration              // Время жизни контекста
}
```

#### `handlers.go` - Обработчики команд
- Обработка команд (/start, /help, /add и т.д.)
- Обработка callback запросов
- Управление рабочими процессами
- Интеграция с сервисами

```go
type BotHandlers struct {
    bot              *Bot                    // Экземпляр бота
    habitService     *service.HabitService   // Сервис привычек
    userService      *service.UserService    // Сервис пользователей
    logService       *service.LogService     // Сервис логирования
    reminderService  *service.ReminderService // Сервис напоминаний
    contextManager   *ContextManager         // Менеджер контекстов
    logger           Logger                  // Логгер
}
```

### 2. Service Layer (`internal/service`)

#### `habit_service.go` - Работа с привычками
- Создание привычек
- Получение информации о привычках
- Обновление привычек
- Установка дней недели/месяца

#### `user_service.go` - Работа с пользователями
- Создание/получение пользователей
- Обновление профиля пользователя
- Интеграция с Telegram ID

#### `log_service.go` - Логирование выполнений
- Логирование выполнения привычек
- Получение истории логов
- Расчет процента выполнения
- Анализ по периодам

#### `reminder_service.go` - Управление напоминаниями
- Генерация напоминаний
- Отметка выполнения
- Запросы по датам и пользователям

### 3. gRPC Client Layer (`internal/grpc`)

#### `client.go` - gRPC клиент
- Подключение к HobitsService
- Управление соединением
- Health checks

```go
type Client struct {
    conn *grpc.ClientConn  // gRPC соединение
}
```

### 4. Configuration & Logging (`internal/config`, `internal/logger`)

#### `config.go` - Конфигурация
- Загрузка переменных окружения
- Конфигурация из .env файла
- Значения по умолчанию

#### `logger.go` - Логирование
- Уровни логирования (DEBUG, INFO, WARN, ERROR)
- Форматирование логов
- Контроль вывода

## Поток данных

### 1. Обработка команды /add

```
User Command (/add)
    ↓
[handleCommand] in main.go
    ↓
handlers.HandleAddHabit()
    ↓
contextManager.SetState("waiting_habit_name")
    ↓
Send message with ForceReply
    ↓
User sends habit name
    ↓
[handleMessage] detects reply
    ↓
handlers.HandleHabitName()
    ↓
contextManager.SetData("habit_name", name)
contextManager.SetState("waiting_frequency")
    ↓
Send frequency keyboard
    ↓
User selects frequency
    ↓
[handleCallback] → handlers.HandleCallbackQuery()
    ↓
handleFrequencySelect()
    ↓
habitService.CreateHabit(ctx, req)
    ↓
gRPC call to HobitsService
    ↓
Response with created habit
    ↓
Send confirmation to user
```

### 2. Обработка callback нажатия кнопки

```
User presses button
    ↓
Callback Query received
    ↓
[handleUpdate] → [handleCallbackQuery]
    ↓
Parse callback data
    ↓
Route to appropriate handler:
    - habit_* → handleHabitAction
    - freq_* → handleFrequencySelect
    - weekday_* → handleWeekdaySelect
    - menu_* → handleMenuSelect
    ↓
Execute action
    ↓
Update message or send new message
```

### 3. Получение списка привычек

```
User command (/habits)
    ↓
handlers.HandleGetHabits()
    ↓
habitService.GetActiveHabits(ctx, userID)
    ↓
gRPC call
    ↓
HobitsService returns habits
    ↓
FormatHabitList()
    ↓
HabitsListKeyboard() with buttons
    ↓
Send formatted message with keyboard
```

## Thread-Safety

### Использование Goroutines

```go
// В main.go
go runBotUpdateLoop(botInstance, handlers, log)

// В runBotUpdateLoop
for update := range updates {
    go handleUpdate(update, ...)  // Каждое обновление в отдельной горутине
}
```

### Синхронизация доступа к контексту

```go
// ContextManager использует sync.RWMutex
type ContextManager struct {
    contexts map[int64]*UserContext
    mu       sync.RWMutex  // Защита от race conditions
}

// Операции захватывают lock
func (cm *ContextManager) SetState(userID int64, state string) {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    // Изменение данных
}
```

## Error Handling

```go
// В handleUpdate()
defer func() {
    if r := recover(); r != nil {
        log.Error("Recovered from panic: %v", r)
    }
}()

// При gRPC вызовах
if err != nil {
    log.Error("failed to get habit: %v", err)
    bot.sendMessage(chatID, FormatError("привычка не найдена"))
    return
}
```

## Performance Considerations

### 1. Connection Pooling
- gRPC соединение переиспользуется для всех запросов
- Отдельное соединение на каждого бота

### 2. Context Management
- TTL (30 минут) автоматически очищает старые контексты
- Background goroutine для очистки
- O(n) сложность очистки, где n - количество активных пользователей

### 3. Message Handling
- Каждое обновление обрабатывается в отдельной горутине
- Нет блокировок при получении новых обновлений
- Длительные операции не влияют на общую производительность

### 4. Logging
- Различные уровни логирования снижают overhead
- Только релевантные логи выводятся

## Security

### 1. Token Management
- Telegram Bot Token хранится в переменной окружения
- Не сохраняется в коде или логах

### 2. User Context
- Привязана к user_id
- Автоматическая очистка по TTL
- Не сохраняется на диск

### 3. Input Validation
- Проверка длины названия привычки
- Проверка на пустые значения
- Валидация дней недели и месяца

### 4. Error Messages
- Безопасные сообщения об ошибках для пользователя
- Детальная информация в логах

## Масштабируемость

### Горизонтальное масштабирование

Для запуска нескольких экземпляров бота:

```yaml
# Kubernetes deployment
replicas: 3
```

Каждый экземпляр:
- Подключается к одному Telegram Bot API
- Имеет собственное gRPC соединение к HobitsService
- Имеет собственный ContextManager (данные не шарятся)

### Вертикальное масштабирование

- Увеличение CPU/Memory для большего количества параллельных обновлений
- Оптимизация gRPC соединения (compression, max connection age)

## Future Improvements

1. **Redis Cache** - Кеширование информации о привычках
2. **Database** - Локальное хранилище для offline режима
3. **Middleware** - Rate limiting, authentication
4. **Metrics** - Prometheus metrics для мониторинга
5. **Testing** - Unit tests и integration tests
6. **Webhook** - Вместо polling для получения обновлений
