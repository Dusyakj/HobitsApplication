# 👋 START HERE - HobitsApplication

**Welcome!** 🎉 Выберите ваш путь:

---

## 🪟 Я на Windows

### ⚡ Способ 1: PowerShell (Рекомендуется)

```powershell
# 1. Создать .env
copy .env.example .env

# 2. Отредактировать (вставить токен)
notepad .env

# 3. Запустить всё
docker-compose up -d

# 4. Проверить статус
docker-compose ps
```

→ **[POWERSHELL_QUICKSTART.md](POWERSHELL_QUICKSTART.md)** - Полное руководство

### ⚡ Способ 2: Batch файлы (Двойной клик)

1. **Двойной клик** → `setup.bat` (создаст .env)
2. **Отредактируйте** → `.env` (вставьте токен)
3. **Двойной клик** → `start.bat` (запустит всё)

→ **[WINDOWS_QUICKSTART.md](WINDOWS_QUICKSTART.md)** - Полное руководство

### 📋 Доступные batch файлы:

```
setup.bat          - Инициализация (первый раз)
start.bat          - Запуск всех сервисов ✓
stop.bat           - Остановка
status.bat         - Статус контейнеров
logs.bat           - Все логи
logs-bot.bat       - Логи бота
logs-service.bat   - Логи backend'а
clean.bat          - Удалить всё (⚠️)
help.bat           - Справка
```

---

## 🐧 Я на Linux / macOS

### ⚡ Быстрый старт:

```bash
# Копировать конфигурацию
cp .env.example .env

# Отредактировать .env и установить TELEGRAM_BOT_TOKEN
nano .env

# Запустить всё
docker-compose up -d

# Проверить статус
docker-compose ps
```

### 📖 Полная документация:
→ **[QUICKSTART.md](QUICKSTART.md)** или **[README.md](README.md)**

---

## 📱 Как получить Telegram Bot Token

### 5 шагов:

1. **Откройте Telegram**
   - Desktop, Web или Mobile

2. **Найдите @BotFather**
   - Поиск → `@BotFather` → Открыть

3. **Создайте бота**
   - Отправьте: `/newbot`
   - Придумайте имя (например: MyHabitsBot)
   - Придумайте username (например: myhabitsbot)

4. **Скопируйте токен**
   - BotFather отправит токен вида:
   ```
   123456789:ABCdefGHIjklmnoPQRstuvWXYZabcdefGHI
   ```
   - Скопируйте полностью (без пробелов!)

5. **Вставьте в `.env`**
   ```
   TELEGRAM_BOT_TOKEN=123456789:ABCdefGHIjklmnoPQRstuvWXYZabcdefGHI
   ```

---

## ✅ После запуска

### Сервисы в браузере:

- **Grafana** (дашборд): http://localhost:3000
  - Логин: `admin`
  - Пароль: `admin`

- **Prometheus** (метрики): http://localhost:9090

- **RabbitMQ** (очередь): http://localhost:15672
  - Логин: `guest`
  - Пароль: `guest`

### Тестирование бота:

1. Откройте Telegram
2. Найдите вашего бота
3. Отправьте `/start`
4. Начните добавлять привычки!

---

## 📚 Документация

### Для быстрого старта:
- **WINDOWS_QUICKSTART.md** - Windows пользователи
- **QUICKSTART.md** - Linux/macOS пользователи

### Для полной информации:
- **README.md** - Полная документация проекта
- **DEPLOYMENT.md** - Production развертывание
- **PROJECT_STRUCTURE.md** - Структура проекта

### Для разработчиков:
- **HobitsBot/README.md** - О боте
- **HobitsBot/ARCHITECTURE.md** - Архитектура
- **HobitsBot/EXAMPLES.md** - Примеры

---

## 🆘 Что если что-то не работает?

### Windows:

Двойной клик → `help.bat`

Или смотрите **[WINDOWS_QUICKSTART.md](WINDOWS_QUICKSTART.md#-частые-проблемы)**

### Linux/macOS:

Смотрите **[DEPLOYMENT.md](DEPLOYMENT.md#troubleshooting)**

---

## 🎯 Архитектура системы

```
Telegram Users
     ↓
  HobitsBot (ваш бот)
     ↓
HobitsService (backend)
     ↓
PostgreSQL (данные)
```

Всё в Docker контейнерах, всё работает автоматически! ✅

---

## 🚀 Готовы начать?

### Windows пользователи:
1. Двойной клик → **setup.bat**
2. Редактируйте → **.env**
3. Двойной клик → **start.bat**
4. Готово! ✓

### Linux/macOS пользователи:
```bash
cp .env.example .env
nano .env  # добавьте токен
docker-compose up -d
```

---

## ❓ Часто задаваемые вопросы

### Q: Где скачать Docker?
A: https://www.docker.com/products/docker-desktop

### Q: Что если батники не запускаются?
A: Смотрите WINDOWS_QUICKSTART.md → Советы Windows пользователям

### Q: Как остановить?
A: Двойной клик → `stop.bat` (Windows) или `docker-compose down` (Linux/macOS)

### Q: Как удалить всё?
A: Двойной клик → `clean.bat` (Windows) ⚠️ это удалит данные!

### Q: Где найти логи ошибок?
A:
- Windows: Двойной клик → `logs.bat`
- Linux/macOS: `docker-compose logs -f`

---

## 📞 Нужна помощь?

1. Прочитайте документацию выше
2. Проверьте логи (help.bat или docker-compose logs)
3. Убедитесь что Docker установлен и запущен
4. Проверьте что `.env` правильно заполнен

---

## 🎓 Что дальше?

1. **Изучите бота** - Используйте все команды в Telegram
2. **Смотрите дашборды** - Откройте Grafana (http://localhost:3000)
3. **Добавляйте привычки** - Начните отслеживание
4. **Читайте документацию** - Узнайте как всё устроено

---

## 📝 Чек-лист:

- [ ] Docker установлен и запущен
- [ ] `.env` создан с вашим токеном
- [ ] Все сервисы запущены (status.bat или docker-compose ps)
- [ ] Бот добавлен в Telegram
- [ ] Отправлен `/start` команде боту
- [ ] Добавлена первая привычка

---

## ✨ Готово!

**Ваша система отслеживания привычек работает!** 🎉

Начните добавлять привычки и отслеживайте прогресс.

**Удачи!** 🚀

---

**Версия:** 1.0
**Дата:** 26 ноября 2024
**Статус:** ✅ Production Ready
