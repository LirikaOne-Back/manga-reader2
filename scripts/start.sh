#!/bin/bash
set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

show_help() {
  echo "Использование: $0 [опции]"
  echo "Опции:"
  echo "  --build          Собрать приложение перед запуском"
  echo "  --docker         Запустить через Docker Compose"
  echo "  --migrate        Выполнить миграции базы данных перед запуском"
  echo "  --dev            Запустить в режиме разработки (с автоматической перезагрузкой)"
  echo "  --help           Показать это сообщение"
}

BUILD=false
DOCKER=false
MIGRATE=false
DEV=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --build)
      BUILD=true
      shift
      ;;
    --docker)
      DOCKER=true
      shift
      ;;
    --migrate)
      MIGRATE=true
      shift
      ;;
    --dev)
      DEV=true
      shift
      ;;
    --help)
      show_help
      exit 0
      ;;
    *)
      echo -e "${RED}Неизвестная опция: $1${NC}"
      show_help
      exit 1
      ;;
  esac
done

if [ -f .env ]; then
  echo -e "${YELLOW}Загрузка переменных окружения из .env файла...${NC}"
  export $(grep -v '^#' .env | xargs)
fi

if [ ! -f .env ] && [ "$DOCKER" = false ]; then
  if [ -f .env.example ]; then
    echo -e "${YELLOW}Файл .env не найден. Копируем из .env.example...${NC}"
    cp .env.example .env
    echo -e "${GREEN}Файл .env создан из примера.${NC}"
    echo -e "${YELLOW}Пожалуйста, проверьте и настройте переменные окружения в .env файле.${NC}"
  else
    echo -e "${RED}Файлы .env и .env.example не найдены.${NC}"
    echo -e "${YELLOW}Создайте файл .env с необходимыми переменными окружения.${NC}"
    exit 1
  fi
fi

check_postgres() {
  echo -e "${YELLOW}Проверка доступности PostgreSQL...${NC}"
  local max_attempts=10
  local attempt=1

  while [ $attempt -le $max_attempts ]; do
    if PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "SELECT 1" &>/dev/null; then
      echo -e "${GREEN}PostgreSQL доступен!${NC}"
      return 0
    fi

    echo -e "${YELLOW}Попытка $attempt/$max_attempts: PostgreSQL недоступен. Ожидаем...${NC}"
    sleep 2
    attempt=$((attempt + 1))
  done

  echo -e "${RED}PostgreSQL недоступен после $max_attempts попыток.${NC}"
  return 1
}

if [ "$DOCKER" = true ]; then
  echo -e "${YELLOW}Запуск приложения через Docker Compose...${NC}"

  if [ "$BUILD" = true ]; then
    echo -e "${YELLOW}Пересборка контейнеров...${NC}"
    docker-compose build
  fi

  echo -e "${YELLOW}Запуск контейнеров...${NC}"
  docker-compose up -d

  if [ "$MIGRATE" = true ]; then
    echo -e "${YELLOW}Ожидание готовности PostgreSQL...${NC}"
    sleep 5
    echo -e "${YELLOW}Выполнение миграций...${NC}"
    docker-compose exec app /app/manga-reader migrate up
  fi

  echo -e "${GREEN}Приложение запущено через Docker Compose.${NC}"
  echo -e "${YELLOW}Для просмотра логов используйте: docker-compose logs -f${NC}"
  echo -e "${YELLOW}Для остановки используйте: docker-compose down${NC}"

  exit 0
fi

if ! check_postgres; then
  echo -e "${RED}PostgreSQL недоступен. Пожалуйста, убедитесь, что сервер запущен и доступен.${NC}"
  echo -e "${YELLOW}Вы можете запустить приложение через Docker Compose с помощью опции --docker${NC}"
  exit 1
fi

if [ "$MIGRATE" = true ]; then
  echo -e "${YELLOW}Выполнение миграций базы данных...${NC}"
  ./scripts/migrate.sh up
fi

if [ "$BUILD" = true ]; then
  echo -e "${YELLOW}Сборка приложения...${NC}"
  go build -o manga-reader ./cmd/api/main.go
  echo -e "${GREEN}Приложение успешно собрано.${NC}"
fi

if [ "$DEV" = true ]; then
  if ! command -v air &> /dev/null; then
    echo -e "${YELLOW}Утилита air не найдена. Устанавливаем...${NC}"
    go install github.com/cosmtrek/air@latest
    echo -e "${GREEN}Утилита air успешно установлена.${NC}"
  fi

  echo -e "${YELLOW}Запуск приложения в режиме разработки...${NC}"
  air
else
  echo -e "${YELLOW}Запуск приложения...${NC}"

  if [ ! -f ./manga-reader ]; then
    echo -e "${YELLOW}Приложение не найдено. Выполняется сборка...${NC}"
    go build -o manga-reader ./cmd/api/main.go
  fi

  ./manga-reader
fi