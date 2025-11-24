# --- Stage 1: Builder ---
# Обновили версию Go до 1.23
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY . .

# Авто-инициализация зависимостей
RUN if [ ! -f go.mod ]; then go mod init core-go; fi && \
    go mod tidy

# Сборка
RUN CGO_ENABLED=0 GOOS=linux go build -o core-app ./cmd/main.go

# --- Stage 2: Production Image ---
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/core-app .

EXPOSE 8080

CMD ["./core-app"]