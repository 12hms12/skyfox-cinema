#!/bin/sh

# This script runs database migrations inside the Docker container

CONTAINER_NAME="${POSTGRES_CONTAINER:-postgresdb}"
DB_NAME="${DB_NAME:-bookingengine}"
DB_USER="${POSTGRES_USERNAME:-bookingengine}"
DB_PASSWORD="${POSTGRES_PASSWORD:-postgres}"

echo "Running database migrations in Docker container: $CONTAINER_NAME"

# Copy migration scripts to the container
docker cp migration/scripts $CONTAINER_NAME:/tmp/

# Run each migration in order
echo "Running migration: 000001_create_slot_table.up.sql"
docker exec -e PGPASSWORD=$DB_PASSWORD $CONTAINER_NAME \
  psql -h localhost -p 5432 -U $DB_USER -d $DB_NAME -f /tmp/scripts/000001_create_slot_table.up.sql

echo "Running migration: 000002_create_show_table.up.sql"
docker exec -e PGPASSWORD=$DB_PASSWORD $CONTAINER_NAME \
  psql -h localhost -p 5432 -U $DB_USER -d $DB_NAME -f /tmp/scripts/000002_create_show_table.up.sql

echo "Running migration: 000003_create_customer_table.up.sql"
docker exec -e PGPASSWORD=$DB_PASSWORD $CONTAINER_NAME \
  psql -h localhost -p 5432 -U $DB_USER -d $DB_NAME -f /tmp/scripts/000003_create_customer_table.up.sql

echo "Running migration: 000004_create_booking_table.up.sql"
docker exec -e PGPASSWORD=$DB_PASSWORD $CONTAINER_NAME \
  psql -h localhost -p 5432 -U $DB_USER -d $DB_NAME -f /tmp/scripts/000004_create_booking_table.up.sql

echo "Running migration: 000005_create_user_table.up.sql"
docker exec -e PGPASSWORD=$DB_PASSWORD $CONTAINER_NAME \
  psql -h localhost -p 5432 -U $DB_USER -d $DB_NAME -f /tmp/scripts/000005_create_user_table.up.sql

echo "All migrations completed successfully!"
