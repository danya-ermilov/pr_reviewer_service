#!/bin/sh
set -e

echo "Waiting for database to be ready..."
until pg_isready -h db -U pruser; do
  sleep 1
done

echo "Database is ready, applying migrations..."
./prreview migrate --database-url "$DATABASE_URL"

echo "Starting application..."
exec ./prreview
