#!/bin/bash

set -e

DB_HOST=localhost
DB_PORT=5432
DB_USER=admin
DB_PASSWORD=password
DB_NAME=basketball_league

for file in ./internal/db/migrations/*.sql; do
    echo "Applying migration: $file"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f $file
done

echo "Migrations applied successfully!"