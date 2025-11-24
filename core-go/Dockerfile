# --- Stage 1: Builder ---
FROM golang:1.21-alpine AS builder

# Устанавливаем необходимые системные пакеты (git нужен для скачивания зависимостей)
RUN apk add --no-cache git

WORKDIR /app

# Сначала копируем файлы зависимостей (для кэширования слоев Docker)
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник
# CGO_ENABLED=0 делает бинарник статическим (работает без внешних библиотек C)
RUN CGO_ENABLED=0 GOOS=linux go build -o core-app ./cmd/main.go

# --- Stage 2: Production Image ---
FROM alpine:latest

# Устанавливаем SSL сертификаты для HTTPS запросов (важно для Telegram/n8n)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Копируем скомпилированный бинарник из первого этапа
COPY --from=builder /app/core-app .

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./core-app"]