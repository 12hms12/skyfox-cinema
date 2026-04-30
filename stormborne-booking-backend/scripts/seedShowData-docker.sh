#!/bin/sh

# This script runs the seeding inside the Docker container to avoid conflicts with local PostgreSQL

CONTAINER_NAME="${POSTGRES_CONTAINER:-db}"
START_DATE="${1:-$(date +%Y-%m-%d)}"
NUM_DAYS="${2:-21}"

echo "Running seed script inside Docker container: $CONTAINER_NAME"
echo "Start date: $START_DATE"
echo "Number of days: $NUM_DAYS"

# Copy the seed script into the container
docker cp scripts/seedShowData.sh $CONTAINER_NAME:/tmp/seedShowData.sh

# Execute the script inside the container
docker exec -e DB_HOST=localhost \
  -e DB_PORT=5432 \
  -e DB_NAME=bookingengine \
  -e POSTGRES_USERNAME=bookingengine \
  -e POSTGRES_PASSWORD=postgres \
  $CONTAINER_NAME \
  sh /tmp/seedShowData.sh "$START_DATE" "$NUM_DAYS"

echo "Seed script completed!"
