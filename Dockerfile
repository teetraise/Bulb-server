# Используем официальный образ Go
FROM golang:1.21-alpine AS builder

# Устанавливаем git (нужен для go mod)
RUN apk add --no-cache git

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go mod и sum файлы
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем весь код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# Финальный образ
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарный файл из builder
COPY --from=builder /app/main .

# Копируем конфиги
COPY --from=builder /app/configs ./configs

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]