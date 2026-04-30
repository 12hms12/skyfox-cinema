#!/usr/bin/env bash
set -euo pipefail

# --- Config (you can export these before running instead of hardcoding) ---
: "${POSTGRES_CONTAINER:=postgresdb}"
: "${DB_NAME:=bookingengine}"
: "${POSTGRES_USERNAME:=bookingengine}"
: "${POSTGRES_PASSWORD:=postgres}"

UP_SQL="migration/scripts/000007_predefined_avatars_and_user_avatar.up.sql"
DOWN_SQL="migration/scripts/000007_predefined_avatars_and_user_avatar.down.sql"

usage() {
  echo "Usage: $0 [up|down]"
  exit 1
}

cmd="${1:-up}"
if [[ "$cmd" != "up" && "$cmd" != "down" ]]; then
  usage
fi

SQL_FILE="$UP_SQL"
[[ "$cmd" == "down" ]] && SQL_FILE="$DOWN_SQL"

echo "Running migration '$cmd' using $SQL_FILE on container '$POSTGRES_CONTAINER' / db '$DB_NAME' as user '$POSTGRES_USERNAME'"

# Run psql inside the container
docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" -i "$POSTGRES_CONTAINER" \
  psql "postgresql://$POSTGRES_USERNAME:$POSTGRES_PASSWORD@localhost:5432/$DB_NAME" \
  -v ON_ERROR_STOP=1 \
  -f - < "$SQL_FILE"
[]
echo "Migration '$cmd' completed successfully."