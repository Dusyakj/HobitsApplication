# 🔵 HobitsApplication - PowerShell Quick Start

**Вы используете PowerShell?** Отлично! Это руководство для вас.

---

## ⚡ Быстрый старт (4 команды)

### 1️⃣ Откройте PowerShell в папке проекта

```powershell
# В Windows 11/10: правый клик в папке HobitsApplication → Open in Terminal
# Или: Win+X → Windows PowerShell

# Убедитесь что вы в правильной папке
cd C:\Users\YourUsername\GolandProjects\HobitsApplication
```

### 2️⃣ Создайте .env файл

```powershell
copy .env.example .env
```

### 3️⃣ Отредактируйте .env

```powershell
# Откройте в notepad
notepad .env
```

**Найдите строку:**
```
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
```

**Замените на ваш токен:**
```
TELEGRAM_BOT_TOKEN=123456789:ABCdefGHIjklmnoPQRstuvWXYZabcdefGHI
```

Сохраните (Ctrl+S) и закройте.

### 4️⃣ Запустите все сервисы

```powershell
docker-compose up -d
```

**Ждите пока завершится (~2 минуты на первый запуск)**

Вы должны увидеть:
```
[+] Running 6/6
  ✔ Container hobits-postgres      Healthy
  ✔ Container hobits-rabbitmq      Healthy
  ✔ Container hobits-service       Started
  ✔ Container hobitsbot            Started
  ✔ Container hobits-prometheus    Started
  ✔ Container hobits-grafana       Started
```

---

## ✅ Проверка статуса

```powershell
docker-compose ps
```

Все контейнеры должны быть в статусе **`running`** или **`healthy`**

---

## 📊 Доступные сервисы

После запуска откройте в браузере:

```
Grafana Dashboard    http://localhost:3000    (admin/admin)
Prometheus           http://localhost:9090
RabbitMQ Manager     http://localhost:15672   (guest/guest)
PostgreSQL           localhost:5432
HobitsService gRPC   localhost:50051
```

---

## 🤖 Тестирование бота

1. Откройте **Telegram**
2. Найдите вашего бота (по имени от BotFather)
3. Отправьте `/start`
4. Используйте команды:
   - `/add` - добавить привычку
   - `/habits` - список привычек
   - `/today` - напоминания
   - `/stats` - статистика
   - `/help` - справка

---

## 📋 Полезные PowerShell команды

### Просмотр логов

```powershell
# Все логи
docker-compose logs -f

# Только бот
docker-compose logs -f hobitsbot

# Только backend
docker-compose logs -f hobits-service

# Только БД
docker-compose logs -f postgres

# Выход из логов: Ctrl+C
```

### Управление сервисами

```powershell
# Статус
docker-compose ps

# Перезагрузить все
docker-compose restart

# Остановить все
docker-compose down

# Полная очистка (удалит данные!)
docker-compose down -v

# Пересобрать образы
docker-compose build --no-cache

# Заново запустить
docker-compose up -d
```

### Работа с БД

```powershell
# Подключиться к PostgreSQL
docker-compose exec postgres psql -U hobits -d hobits

# Примеры команд SQL:
# SELECT * FROM users;
# SELECT * FROM habits;
# \dt                    - показать все таблицы
# \q                     - выход
```

---

## 🆘 Частые проблемы

### Проблема: "docker: command not found"

**Решение:**
1. Убедитесь что Docker Desktop установлен
2. Docker должен быть запущен (проверьте системный трей)
3. Перезагрузитесь
4. Попробуйте снова

### Проблема: "Port 5432 is already in use"

**Решение:**
```powershell
# Остановите старые контейнеры
docker-compose down

# Подождите 10 секунд и запустите снова
docker-compose up -d
```

### Проблема: Бот не реагирует

**Решение:**
```powershell
# 1. Проверьте логи бота
docker-compose logs hobitsbot

# 2. Убедитесь что токен правильный в .env
notepad .env

# 3. Проверьте что все контейнеры running
docker-compose ps

# 4. Перезагрузите
docker-compose restart hobitsbot
```

### Проблема: Нужно всё переделать с нуля

**Решение:**
```powershell
# Удалите всё
docker-compose down -v

# Удалите .env
Remove-Item .env

# Создайте новый .env
copy .env.example .env

# Отредактируйте и запустите
notepad .env
docker-compose up -d
```

---

## 💡 PowerShell советы

### 1. Создайте alias для удобства

```powershell
# Добавьте в PowerShell profile (опционально):
Set-Alias dcps 'docker-compose ps'
Set-Alias dcl 'docker-compose logs'
Set-Alias dcu 'docker-compose up -d'
Set-Alias dcd 'docker-compose down'
```

### 2. Используйте Clear для чистоты

```powershell
# Очистить экран
Clear-Host

# Или просто
cls
```

### 3. Скопируйте и вставляйте с Shift+Ins

```
Ctrl+V - не всегда работает в PowerShell
Shift+Ins - стандартный способ вставки
```

---

## 🚀 Полный цикл разработки

```powershell
# 1. Создание и конфигурация
copy .env.example .env
notepad .env

# 2. Запуск
docker-compose up -d

# 3. Проверка
docker-compose ps
docker-compose logs -f hobitsbot

# 4. Тестирование в Telegram
# (отправьте /start боту)

# 5. Просмотр дашборда
# http://localhost:3000 в браузере

# 6. Остановка
docker-compose down

# 7. Полная очистка (если нужно переделать)
docker-compose down -v
```

---

## 📞 Быстрая справка

```powershell
# Когда вы забыли команду, используйте:
docker-compose --help

# Когда контейнер не запускается:
docker-compose logs service_name

# Когда порт занят:
docker-compose down && docker-compose up -d

# Когда нужно рестартнуть:
docker-compose restart

# Когда всё сломано:
docker-compose down -v
```

---

## 🎓 Команды бота в Telegram

Используйте эти команды в чате с ботом:

```
/start    - Инициализация и приветствие
/add      - Добавить новую привычку
/habits   - Показать список всех привычек
/today    - Напоминания на сегодня
/stats    - Статистика по привычкам
/help     - Справка по всем командам
```

---

## ✨ Готово!

Ваша система полностью настроена и работает!

**Рекомендуемые дальнейшие шаги:**

1. Добавьте 2-3 привычки через `/add`
2. Отмечайте выполнение через `/today`
3. Смотрите статистику в Grafana (http://localhost:3000)
4. Читайте полную документацию в README.md

---

## 📚 Дополнительная документация

- **START_HERE.md** - Общее руководство
- **README.md** - Полная документация
- **DEPLOYMENT.md** - Production развертывание
- **HobitsBot/EXAMPLES.md** - Примеры использования

---

**Happy habits tracking!** 🎉

Удачи в отслеживании привычек! 🚀
