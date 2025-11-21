# Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем go mod файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /server cmd/server/main.go

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем бинарник
COPY --from=builder /server .

# Копируем миграции
COPY migrations ./migrations

EXPOSE 50051

CMD ["./server"]
