# Этап сборки
FROM golang:1.24-alpine AS builder

# Установка необходимых зависимостей
RUN apk add --no-cache git gcc musl-dev

# Установка рабочей директории
WORKDIR /app

# Копирование файлов go.mod и go.sum
COPY go.mod go.sum ./

# Загрузка зависимостей
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=1 GOOS=linux go build -a -o manga-reader ./cmd/api/main.go

# Этап запуска
FROM alpine:latest

# Установка необходимых пакетов
RUN apk add --no-cache ca-certificates tzdata

# Создание пользователя без привилегий
RUN adduser -D -u 1000 appuser

# Создание директорий для загруженных файлов и данных
RUN mkdir -p /app/uploads /app/data && \
    chown -R appuser:appuser /app

WORKDIR /app

# Копирование исполняемого файла из этапа сборки
COPY --from=builder /app/manga-reader /app/
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/.env.example /app/.env

# Назначение прав на исполнение
RUN chmod +x /app/manga-reader

# Переключение на пользователя без привилегий
USER appuser

# Установка переменных окружения
ENV SERVER_ADDRESS=:8080

# Открытие порта
EXPOSE 8080

# Запуск приложения
CMD ["/app/manga-reader"]