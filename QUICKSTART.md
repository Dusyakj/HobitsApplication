# ⚡ Быстрый старт HobitsApplication

**Время установки:** ~5 минут
**Требования:** Docker и Docker Compose

## Шаг 1️⃣: Подготовка

```bash
# Перейти в папку проекта
cd HobitsApplication

# Скопировать пример конфигурации
cp .env.example .env
```

## Шаг 2️⃣: Конфигурация Telegram Bot

1. Откройте файл `.env`
2. Найдите строку `TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here`
3. Замените на ваш реальный токен

**Как получить токен?**
- Откройте Telegram
- Найдите `@BotFather`
- Отправьте `/newbot`
- Следуйте инструкциям
- Скопируйте токен

**Пример .env:**
```env
TELEGRAM_BOT_TOKEN=123456789:ABCdefGHIjklmnoPQRstuvWXYZabcdefGHI

# Остальное с дефолтными значениями
DB_USER=hobits
DB_PASSWORD=password
DB_NAME=hobits
GRAFANA_USER=admin
GRAFANA_PASSWORD=admin
```

## Шаг 3️⃣: Запуск всех сервисов

```bash
# Запустить все контейнеры
docker-compose up -d

# Проверить статус
docker-compose ps
```

Вы должны увидеть все сервисы в статусе `running`:
```
NAME                    STATUS
hobits-postgres         running
hobits-rabbitmq         running
hobits-service          running
hobitsbot               running
hobits-prometheus       running
hobits-grafana          running
```

## Шаг 4️⃣: Тестирование бота

1. Откройте Telegram
2. Найдите своего бота по имени (которое вы выбрали при создании в BotFather)
3. Отправьте команду `/start`

**Ожидаемый ответ:**
```
👋 Добро пожаловать!

Я ваш личный помощник для отслеживания привычек. 🎯

Со мной вы сможете:
✅ Создавать и отслеживать свои привычки
📊 Анализировать прогресс
🔔 Получать напоминания
🏆 Видеть статистику и достижения

Начнём? Создайте первую привычку или посмотрите справку.
```

## 📊 Доступные сервисы

| Сервис | URL | Логин | Пароль |
|--------|-----|-------|--------|
| **Grafana** (Дашборд) | http://localhost:3000 | admin | admin |
| **Prometheus** (Метрики) | http://localhost:9090 | - | - |
| **RabbitMQ** (Очередь) | http://localhost:15672 | guest | guest |
| **PostgreSQL** (БД) | localhost:5432 | hobits | password |

## 🤖 Команды бота

```
/start    - Начало работы
/habits   - Мои привычки
/add      - Добавить привычку
/today    - Напоминания на сегодня
/stats    - Статистика привычек
/help     - Справка по командам
```

## 📝 Пример: Добавление привычки

1. Отправьте `/add`
2. Введите название: "Зарядка"
3. Выберите частоту: "Ежедневно"
4. Подтвердите

✅ Привычка создана! Бот будет напоминать вам о ней.

## 🛠️ Полезные команды

### Просмотр логов

```bash
# Все сервисы
docker-compose logs -f

# Конкретный сервис
docker-compose logs -f hobitsbot      # Бот
docker-compose logs -f hobits-service # Backend
docker-compose logs -f postgres       # БД
```

### Управление

```bash
# Остановить
docker-compose down

# Перезагрузить
docker-compose restart

# Полная очистка (включая данные!)
docker-compose down -v
```

### Работа с БД

```bash
# Подключиться к PostgreSQL
docker-compose exec postgres psql -U hobits -d hobits

# Пример SQL команд:
# \dt                    - показать таблицы
# SELECT * FROM users;   - показать пользователей
# \q                     - выход
```

## ✅ Проверка установки

Используйте Make команды для удобства:

```bash
make help      # Показать все команды
make setup     # Инициальная конфигурация
make up        # Запустить
make ps        # Статус контейнеров
make logs      # Просмотр логов
make down      # Остановить
```

## 🐛 Если что-то не работает

### Бот не реагирует

```bash
# 1. Проверьте логи бота
docker-compose logs hobitsbot

# 2. Убедитесь в токене (нет пробелов в начале/конце)
grep TELEGRAM_BOT_TOKEN .env

# 3. Проверьте статус сервиса
docker-compose ps hobits-service
```

### Сервис не запускается

```bash
# 1. Просмотрите ошибку
docker-compose logs hobits-service

# 2. Проверьте подключение к БД
docker-compose logs postgres

# 3. Пересоберите образ
docker-compose build --no-cache hobits-service
docker-compose up -d hobits-service
```

### Ошибка при запуске docker-compose

```bash
# Проверьте конфигурацию
docker-compose config

# Убедитесь что .env существует
ls -la .env

# Обновите docker-compose
docker-compose version
```

## 🚀 Следующие шаги

1. **Добавьте несколько привычек** в Telegram боте
2. **Отслеживайте прогресс** через `/today` и `/stats`
3. **Смотрите дашборды** в Grafana (http://localhost:3000)
4. **Изучите документацию** в [README.md](README.md)

## 📚 Дополнительно

- **Архитектура бота**: [HobitsBot/ARCHITECTURE.md](HobitsBot/ARCHITECTURE.md)
- **Примеры использования**: [HobitsBot/EXAMPLES.md](HobitsBot/EXAMPLES.md)
- **Развертывание**: [DEPLOYMENT.md](DEPLOYMENT.md)
- **Полная документация**: [README.md](README.md)

---

**🎉 Готово!** Ваша система отслеживания привычек работает!

Если возникают вопросы, проверьте логи или прочитайте документацию.
