#!/bin/bash
# Danger: This will wipe all tables in the database! Use only for local/dev/test.

set -euo pipefail

# Customize these for your environment
DB_NAME="mental_health_db"
DB_USER="postgres"
DB_HOST="localhost"
DB_PORT="5432"

# Drop all tables and re-run migrations (Postgres example)
psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT" -d "$DB_NAME" -c 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;'

# Optionally, re-run Go migrations if you have them:
# go run main.go migrate

echo "✅ Database cleaned."
