#!/bin/bash
set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

if [ -f .env ]; then
  echo -e "${YELLOW}Загрузка переменных окружения из .env файла...${NC}"
  export $(grep -v '^#' .env | xargs)
fi

check_migrate() {
  if ! command -v migrate &> /dev/null; then
    echo -e "${YELLOW}Утилита migrate не найдена. Устанавливаем...${NC}"

    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $ARCH in
      x86_64) ARCH="amd64" ;;
      aarch64) ARCH="arm64" ;;
      arm*) ARCH="arm64" ;;
    esac

    MIGRATE_VERSION="v4.16.2"
    DOWNLOAD_URL="https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.${OS}-${ARCH}.tar.gz"

    echo -e "${YELLOW}Скачиваем migrate для ${OS}-${ARCH}...${NC}"
    curl -L ${DOWNLOAD_URL} -o migrate.tar.gz
    tar -xzf migrate.tar.gz
    chmod +x migrate

    if [ "$OS" = "darwin" ]; then
      mv migrate /usr/local/bin/
    else
      sudo mv migrate /usr/local/bin/
    fi

    rm migrate.tar.gz
    echo -e "${GREEN}Утилита migrate успешно установлена.${NC}"
  fi
}

COMMAND="up"
DATABASE_URL=""

while [[ $# -gt 0 ]]; do
  case $1 in
    up|down|create|version|force)
      COMMAND="$1"
      shift
      ;;
    -u|--url)
      DATABASE_URL="$2"
      shift
      shift
      ;;
    *)
      shift
      ;;
  esac
done

check_migrate

if [ -z "$DATABASE_URL" ]; then
  if [ -z "$POSTGRES_HOST" ] || [ -z "$POSTGRES_PORT" ] || [ -z "$POSTGRES_USER" ] || [ -z "$POSTGRES_DB" ]; then
    echo -e "${RED}Не указаны переменные окружения для подключения к базе данных.${NC}"
    echo -e "${YELLOW}Используйте .env файл или укажите URL базы данных с помощью параметра -u/--url.${NC}"
    exit 1
  fi

  POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-}
  POSTGRES_SSLMODE=${POSTGRES_SSLMODE:-disable}

  DATABASE_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_SSLMODE}"
fi

echo -e "${YELLOW}Используем URL базы данных: ${DATABASE_URL}${NC}"

MIGRATIONS_DIR="./migrations"

if [ "$COMMAND" = "create" ]; then
  if [ -z "$2" ]; then
    echo -e "${RED}Не указано название миграции.${NC}"
    echo -e "${YELLOW}Использование: $0 create <название_миграции>${NC}"
    exit 1
  fi

  MIGRATION_NAME="$2"
  echo -e "${YELLOW}Создание новой миграции ${MIGRATION_NAME}...${NC}"
  migrate create -ext sql -dir ${MIGRATIONS_DIR} -seq ${MIGRATION_NAME}
else
  echo -e "${YELLOW}Выполнение миграции ${COMMAND}...${NC}"
  migrate -database "${DATABASE_URL}" -path ${MIGRATIONS_DIR} ${COMMAND}
fi

if [ $? -eq 0 ]; then
  echo -e "${GREEN}Миграция успешно выполнена!${NC}"
else
  echo -e "${RED}Произошла ошибка при выполнении миграции.${NC}"
  exit 1
fi