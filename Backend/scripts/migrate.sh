#!/bin/bash

set -e

DB_HOST=localhost
DB_PORT=5432
DB_USER=admin
DB_PASSWORD=password
DB_NAME=basketball_league

# Массив файлов миграций
files=(
    ./internal/db/migrations/0002_create_teams.sql
    ./internal/db/migrations/0001_create_players.sql
    ./internal/db/migrations/0003_create_matches.sql
    ./internal/db/migrations/0004_create_match_statistics.sql
)

for file in "${files[@]}"; do
    echo "Applying migration: $file"

    # Выполнение миграции с игнорированием ошибок существования
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f $file 2>&1 | grep -v "already exists" || true
done

echo "Migrations applied successfully!"
