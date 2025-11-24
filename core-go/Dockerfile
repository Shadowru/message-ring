# --- Stage 1: Builder ---
FROM golang:1.21-alpine AS builder

# Устанавливаем git (нужен для скачивания зависимостей)
RUN apk add --no-cache git

WORKDIR /app

# 1. Копируем ВЕСЬ исходный код сразу
COPY . .

# 2. Магия авто-инициализации:
# Если файла go.mod нет -> создаем его.
# Затем запускаем go mod tidy, который скачает библиотеки и создаст go.sum
RUN if [ ! -f go.mod ]; then go mod init core-go; fi && \
    go mod tidy

# 3. Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o core-app ./cmd/main.go

# --- Stage 2: Production Image ---
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/core-app .

EXPOSE 8080

CMD ["./core-app"]