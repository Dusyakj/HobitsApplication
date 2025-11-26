# 🪟 HobitsApplication - Windows Quick Start

**Время установки:** ~5 минут
**Требования:** Docker Desktop (Windows)

## 🚀 Самый быстрый способ (Клик-клик-готово!)

### Шаг 1️⃣: Выполнить setup.bat

1. Откройте папку `HobitsApplication`
2. Двойной клик на **`setup.bat`**
3. Появится сообщение о создании файла `.env`

### Шаг 2️⃣: Отредактировать .env

1. В папке найдите файл **`.env`**
2. Откройте его с помощью Notepad (правый клик → Открыть с → Notepad)
3. Найдите строку:
   ```
   TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
   ```
4. Замените на ваш токен (см. ниже как его получить)
5. Сохраните файл (Ctrl+S)

### Шаг 3️⃣: Запустить start.bat

1. Двойной клик на **`start.bat`**
2. Дождитесь завершения (обычно 1-2 минуты)
3. Увидите список доступных сервисов

### Шаг 4️⃣: Тестирование бота

1. Откройте Telegram
2. Найдите вашего бота по имени
3. Отправьте `/start`
4. Готово! 🎉

---

## 📱 Как получить Telegram Bot Token

1. **Откройте Telegram** (desktop или web: web.telegram.org)
2. **Найдите @BotFather**
   - В поиске напечатайте: `@BotFather`
   - Откройте бот
3. **Создайте нового бота**
   - Отправьте: `/newbot`
   - Укажите имя бота (например: `MyHabitsBot`)
   - Укажите username (например: `myhabitsbot`)
4. **Скопируйте токен**
   - BotFather отправит вам токен вида: `123456789:ABCdefGHIjklmnoPQRstuvWXYZabcdefGHI`
   - Скопируйте его полностью (без пробелов!)
5. **Вставьте в .env**
   - Откройте `.env` файл
   - Вставьте токен вместо `your_telegram_bot_token_here`

---

## 🎮 Команды для Windows (двойной клик = запуск)

| Файл | Назначение |
|------|-----------|
| **setup.bat** | Создать .env файл (только в первый раз) |
| **start.bat** | Запустить все сервисы ✅ |
| **stop.bat** | Остановить все сервисы |
| **status.bat** | Показать статус контейнеров |
| **logs.bat** | Просмотреть логи всех сервисов |
| **logs-bot.bat** | Просмотреть логи бота |
| **logs-service.bat** | Просмотреть логи backend'а |
| **clean.bat** | Удалить все контейнеры и данные (⚠️ осторожно!) |
| **help.bat** | Показать справку |

---

## 📊 Что запускается после start.bat

### Веб интерфейсы (откройте в браузере):

| Сервис | Адрес | Логин | Пароль |
|--------|-------|-------|--------|
| **Grafana** (Дашборд) | http://localhost:3000 | admin | admin |
| **Prometheus** (Метрики) | http://localhost:9090 | - | - |
| **RabbitMQ** (Очередь) | http://localhost:15672 | guest | guest |

### Для разработчиков:

- **PostgreSQL**: `localhost:5432` (пользователь: `hobits`)
- **HobitsService**: `localhost:50051` (gRPC)
- **REST API**: `http://localhost:8080`

---

## 🤖 Командно бота в Telegram

Когда бот запущен, используйте эти команды:

```
/start    - Начало работы
/habits   - Мои привычки
/add      - Добавить новую привычку
/today    - Напоминания на сегодня
/stats    - Статистика привычек
/help     - Справка по всем командам
```

**Пример:** Добавление привычки
1. Отправьте: `/add`
2. Введите: `Зарядка`
3. Выберите частоту: "Ежедневно"
4. Подтвердите
5. ✅ Привычка добавлена!

---

## 🔍 Проверка установки Docker

Если `start.bat` выдает ошибку, проверьте Docker:

### Вариант 1: PowerShell (рекомендуется)

1. Откройте **PowerShell** (правый клик на рабочем столе → PowerShell)
2. Введите:
   ```powershell
   docker --version
   docker-compose --version
   ```
3. Если увидите версии - Docker установлен ✓

### Вариант 2: Command Prompt

1. Откройте **Command Prompt** (Win+R → cmd → Enter)
2. Введите:
   ```cmd
   docker --version
   docker-compose --version
   ```

### Если Docker не установлен:

1. Скачайте [Docker Desktop для Windows](https://www.docker.com/products/docker-desktop)
2. Установите (запустите скачанный файл)
3. Перезагрузитесь
4. Попробуйте снова

---

## 🐛 Частые проблемы

### Проблема: "docker-compose: command not found"

**Решение:**
- Docker Desktop установлен?
  → Если нет, установите: https://www.docker.com/products/docker-desktop
- Docker запущен?
  → Откройте Docker Desktop (должно быть в системном трее)
- Перезагружали после установки?
  → Перезагрузитесь и попробуйте снова

### Проблема: Бот не отвечает

**Решение:**
1. Проверьте статус: двойной клик `status.bat`
2. Все контейнеры должны быть в статусе `running`
3. Проверьте логи: `logs-bot.bat`
4. Убедитесь, что токен правильный в `.env`

### Проблема: "Error: Port 5432 is already in use"

**Решение:**
- Другой сервис использует порт
- Запустите: `stop.bat`
- Дождитесь завершения
- Попробуйте `start.bat` снова

### Проблема: Нужно перезагрузить

**Решение:**
1. Двойной клик `stop.bat`
2. Дождитесь завершения
3. Двойной клик `start.bat`

### Проблема: Нужно удалить всё и начать заново

⚠️ **Внимание: это удалит все данные!**

1. Двойной клик `clean.bat`
2. Подтвердите (введите Y)
3. Двойной клик `start.bat`

---

## 📚 Дополнительная информация

### Основная документация
- **README.md** - Полная информация о проекте
- **QUICKSTART.md** - Универсальный быстрый старт
- **DEPLOYMENT.md** - Развертывание в production

### Для разработчиков
- **HobitsBot/README.md** - Документация бота
- **HobitsBot/ARCHITECTURE.md** - Архитектура бота
- **HobitsBot/EXAMPLES.md** - Примеры использования

---

## 💡 Советы Windows пользователям

### Используйте PowerShell вместо Command Prompt
```powershell
# PowerShell красивее и функциональнее
# Открыть: Win+X → Windows PowerShell или PowerShell
```

### Добавьте ярлыки на рабочий стол
1. Кликните правой кнопкой на `start.bat`
2. Отправить → Рабочий стол (создать ярлык)
3. Теперь можно запускать прямо с рабочего стола

### Если батники не запускаются
1. Правый клик на файл `.bat`
2. Открыть с → Command Prompt
3. Или измените расширение файла с `.bat` на `.cmd`

---

## ✅ Успешный запуск выглядит так:

```
╔═══════════════════════════════════════════════════════════════╗
║              Starting HobitsApplication                       ║
╚═══════════════════════════════════════════════════════════════╝

[INFO] Starting all services with docker-compose...

Creating network "HobitsApplication_hobits-network" with driver "bridge"
Creating hobits-postgres ... done
Creating hobits-rabbitmq ... done
Creating hobits-service ... done
Creating hobitsbot ... done
Creating hobits-prometheus ... done
Creating hobits-grafana ... done

✓ All services started successfully!

📊 Available services:
   - Grafana Dashboard:  http://localhost:3000 (admin/admin)
   - Prometheus:         http://localhost:9090
   - RabbitMQ Manager:   http://localhost:15672 (guest/guest)
   - PostgreSQL:         localhost:5432
   - HobitsService:      localhost:50051 (gRPC)

🤖 Next steps:
   1. Wait 30 seconds for services to fully start
   2. Open Telegram and find your bot
   3. Send /start command
```

---

## 🎉 Готово!

Ваша система отслеживания привычек запущена и работает!

**Дальше:**
1. Откройте Telegram
2. Найдите вашего бота
3. Отправьте `/start`
4. Начните отслеживать привычки!

---

**Вопросы?** Читайте `help.bat` или смотрите полную документацию в `README.md`

🚀 Удачи!
