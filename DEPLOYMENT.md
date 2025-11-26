# 🚀 Руководство по развертыванию HobitsApplication

## Содержание
1. [Локальная разработка](#локальная-разработка)
2. [Запуск с Docker Compose](#запуск-с-docker-compose)
3. [Запуск на Kubernetes](#запуск-на-kubernetes)
4. [Production развертывание](#production-развертывание)
5. [Мониторинг и логирование](#мониторинг-и-логирование)

---

## Локальная разработка

### Предварительные требования
- Go 1.24+
- PostgreSQL 15+
- RabbitMQ 3.x
- Git

### Установка

1. **Клонируйте репозиторий**
```bash
git clone <repository-url>
cd HobitsApplication
```

2. **Установите зависимости**
```bash
# HobitsService
cd HobitsService
go mod download
cd ../

# HobitsBot
cd HobitsBot
go mod download
cd ../
```

3. **Создайте локальную конфигурацию**
```bash
cp .env.example .env
```

Отредактируйте `.env`:
```env
TELEGRAM_BOT_TOKEN=your_token_here
DB_USER=hobits
DB_PASSWORD=password
DB_NAME=hobits
LOG_LEVEL=debug
```

4. **Запустите PostgreSQL и RabbitMQ локально**

Вариант 1: С Docker
```bash
docker run -d --name postgres -e POSTGRES_PASSWORD=password -p 5432:5432 postgres:15
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management
```

Вариант 2: Установленные локально
```bash
# Убедитесь, что PostgreSQL и RabbitMQ запущены на локальной машине
```

5. **Запустите сервисы**

Terminal 1: HobitsService
```bash
cd HobitsService
make run
# Слушает на :50051 (gRPC) и :8080 (REST)
```

Terminal 2: HobitsBot
```bash
cd HobitsBot
make run
# Подключается к gRPC серверу
```

---

## Запуск с Docker Compose

### Быстрый старт (3 команды)

```bash
# 1. Конфигурация
cp .env.example .env
# Отредактируйте .env и установите TELEGRAM_BOT_TOKEN

# 2. Запуск
docker-compose up -d

# 3. Проверка
docker-compose ps
```

### Доступ к сервисам

| Сервис | URL | Логин | Пароль |
|--------|-----|-------|--------|
| Grafana | http://localhost:3000 | admin | admin |
| Prometheus | http://localhost:9090 | - | - |
| RabbitMQ | http://localhost:15672 | guest | guest |
| PostgreSQL | localhost:5432 | hobits | password |
| HobitsService | localhost:50051 (gRPC) | - | - |
| Backend API | http://localhost:8080 | - | - |

### Управление

```bash
# Просмотр логов
docker-compose logs -f hobitsbot

# Перезагрузка
docker-compose restart

# Остановка
docker-compose down

# Полная очистка
docker-compose down -v
```

### Использование Makefile

```bash
make help         # Показать все команды
make setup        # Инициальная конфигурация
make up           # Запустить сервисы
make down         # Остановить сервисы
make logs         # Просмотреть логи
make db-shell     # Подключиться к БД
```

---

## Запуск на Kubernetes

### Предварительные требования
- Kubernetes 1.20+
- kubectl
- Helm (опционально)

### Структура файлов

```
k8s/
├── namespace.yml
├── postgres/
│   ├── deployment.yml
│   ├── service.yml
│   └── pvc.yml
├── rabbitmq/
│   ├── deployment.yml
│   └── service.yml
├── hobits-service/
│   ├── deployment.yml
│   ├── service.yml
│   └── configmap.yml
├── hobitsbot/
│   ├── deployment.yml
│   └── configmap.yml
├── monitoring/
│   ├── prometheus-deployment.yml
│   └── grafana-deployment.yml
└── ingress.yml
```

### Развертывание

1. **Создайте namespace**
```bash
kubectl create namespace hobits
```

2. **Создайте secret с чувствительными данными**
```bash
kubectl create secret generic hobits-secrets \
  --from-literal=telegram-token=your_token \
  --from-literal=db-password=your_password \
  -n hobits
```

3. **Развертните сервисы**
```bash
kubectl apply -f k8s/ -n hobits
```

4. **Проверьте статус**
```bash
kubectl get pods -n hobits
kubectl get svc -n hobits
```

5. **Настройте Ingress (опционально)**
```bash
# Отредактируйте k8s/ingress.yml с вашим доменом
kubectl apply -f k8s/ingress.yml -n hobits
```

### Масштабирование

```bash
# Увеличить реплики бота
kubectl scale deployment hobitsbot --replicas=3 -n hobits

# Автоматическое масштабирование
kubectl autoscale deployment hobitsbot --min=2 --max=10 -n hobits
```

### Просмотр логов

```bash
# Все логи
kubectl logs -f deployment/hobitsbot -n hobits

# Последние 100 строк
kubectl logs -f deployment/hobitsbot -n hobits --tail=100

# С отметками времени
kubectl logs -f deployment/hobitsbot -n hobits --timestamps=true
```

---

## Production развертывание

### Рекомендации безопасности

1. **Переменные окружения**
   - Используйте Kubernetes Secrets или управление секретами
   - Никогда не коммитьте реальные значения
   - Используйте отдельные значения для prod/dev

2. **Database**
   - Включите SSL для PostgreSQL
   - Используйте strong пароли (15+ символов)
   - Регулярно создавайте резервные копии
   - Шифруйте данные в покое

3. **Telegram Bot Token**
   - Используйте разные токены для dev/prod
   - Храните в защищенном хранилище
   - Ротируйте токены регулярно

4. **Сетевая безопасность**
   - Используйте TLS/SSL для всех соединений
   - Ограничьте доступ к портам firewall'ом
   - Используйте VPN для управления сервисами
   - Настройте Rate Limiting на nginx/ingress

5. **Мониторинг**
   - Включите логирование всех операций
   - Настройте алерты на критические ошибки
   - Регулярно проверяйте метрики производительности

### Production Docker Compose

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/db_password
    secrets:
      - db_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: always

  hobits-service:
    image: registry.example.com/hobits-service:latest
    environment:
      ENVIRONMENT: production
      LOG_LEVEL: warn
    restart: always

  hobitsbot:
    image: registry.example.com/hobitsbot:latest
    environment:
      LOG_LEVEL: warn
    secrets:
      - telegram_token
    restart: always

secrets:
  db_password:
    external: true
  telegram_token:
    external: true

volumes:
  postgres_data:
```

Запуск:
```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Резервные копии

```bash
# Ежедневный backup
docker-compose exec -T postgres pg_dump -U hobits -d hobits | gzip > backup_$(date +%Y%m%d).sql.gz

# Restore
gunzip < backup_20240101.sql.gz | docker-compose exec -T postgres psql -U hobits -d hobits
```

---

## Мониторинг и логирование

### Prometheus Метрики

Доступны на: `http://localhost:9090`

**Ключевые метрики:**
- `hobits_habits_total` - всего привычек
- `hobits_completions_total` - всего выполнений
- `hobits_user_active` - активных пользователей
- `grpc_requests_total` - gRPC запросы
- `http_requests_total` - HTTP запросы

### Grafana Dashboard

Доступен на: `http://localhost:3000`

**Панели:**
- Главная статистика
- Активность пользователей
- Выполнение привычек
- Производительность системы
- Ошибки и alerts

### Elasticsearch & Kibana (опционально)

Для централизованного логирования:

```yaml
elasticsearch:
  image: docker.elastic.co/elasticsearch/elasticsearch:8.0.0
  environment:
    - discovery.type=single-node

kibana:
  image: docker.elastic.co/kibana/kibana:8.0.0
  ports:
    - "5601:5601"
```

### Структурированное логирование

```go
// Пример логирования в коде
logger.Info("habit_created",
  "user_id", userID,
  "habit_id", habitID,
  "frequency", frequency,
)
```

---

## Troubleshooting

### Проблема: Бот не подключается к HobitsService

```bash
# Проверка сетевого подключения
docker-compose exec hobitsbot ping hobits-service

# Проверка логов сервиса
docker-compose logs hobits-service | grep -i error

# Проверка переменных окружения
docker-compose exec hobitsbot env | grep GRPC
```

### Проблема: Ошибка БД при миграции

```bash
# Проверка статуса PostgreSQL
docker-compose ps postgres

# Просмотр логов миграции
docker-compose logs hobits-service | grep -i migration

# Ручное применение миграций
docker-compose exec hobits-service migrate -path ./migrations -database "postgres://..." up
```

### Проблема: Высокое использование памяти

```bash
# Просмотр потребления ресурсов
docker stats

# Проверка утечек памяти
docker-compose logs hobits-service | grep -i "memory\|leak"

# Ограничение ресурсов в docker-compose.yml
services:
  hobits-service:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

---

## Обновление приложения

### Zero-downtime deployment

1. **Соберите новый образ**
```bash
docker build -t hobitsbot:v2.0 ./HobitsBot
docker build -t hobits-service:v2.0 ./HobitsService
```

2. **Загрузите на registry**
```bash
docker push registry.example.com/hobitsbot:v2.0
docker push registry.example.com/hobits-service:v2.0
```

3. **Обновите в Kubernetes**
```bash
kubectl set image deployment/hobitsbot hobitsbot=registry.example.com/hobitsbot:v2.0 -n hobits
kubectl set image deployment/hobits-service hobits-service=registry.example.com/hobits-service:v2.0 -n hobits

# Проверьте rollout
kubectl rollout status deployment/hobitsbot -n hobits
```

---

## Поддержка

Если у вас возникли проблемы:

1. Проверьте логи: `docker-compose logs`
2. Прочитайте документацию в `README.md`
3. Создайте issue с описанием проблемы

---

**Готово к development и production!** 🎉
