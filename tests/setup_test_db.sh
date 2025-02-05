#!/bin/bash
set -e

DB_CONTAINER_NAME="test_postgres_db"
DB_NAME="testdb"
DB_USER="postgres"
DB_PASS="test"
DB_PORT=5433

if [ ! "$(docker ps -q -f name=$DB_CONTAINER_NAME)" ]; then
    echo "Запуск тестовой PostgreSQL..."
    docker run --rm --name $DB_CONTAINER_NAME -e POSTGRES_USER=$DB_USER -e POSTGRES_PASSWORD=$DB_PASS -e POSTGRES_DB=$DB_NAME -p $DB_PORT:5432 -d postgres
    sleep 3
fi

echo "Тестовая база запущена."
