#!/bin/sh

MOVIE_IDS="tt6644200 tt6857112 tt7784604 tt5052448 tt1396484 tt5968394 tt4972582 tt6823368 tt7556122 tt1179933 \
           tt7349950 tt2935510 tt0437086 tt4154664 tt3016748 tt3513498 tt4178092 tt4154796 tt2382320 tt1833116 \
           tt0448115 tt4160708 tt0837563 tt2283336 tt2274648 tt7959026 tt2527338 tt4913966 tt5052474 tt3741700 \
           tt6565702 tt7286456 tt6751668 tt1375666 tt0137523 tt0110912 tt6966692 tt5027774 tt2119532 tt8936646 \
           tt7131622 tt8579674 tt8946378 tt2584384 tt1392190 tt0133093 tt3281548 tt7653254 tt8404614 tt1302006"

slot_ids=""
screen_id=""
start_date=""
next_date=""

# Default values from config-local.yml
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_DATABASE="${DB_NAME:-bookingengine}"
DB_USERNAME="${POSTGRES_USERNAME:-bookingengine}"
DB_PASSWORD="${POSTGRES_PASSWORD:-postgres}"

OS_TYPE=$(uname -s)

get_random_movie_id () {
  echo $MOVIE_IDS | tr ' ' '\n' | shuf -n 1
}

get_random_price_with_two_decimal_places () {
  price_lower_value=150
  price_upper_value=300
  price=$((RANDOM % (price_upper_value - price_lower_value + 1) + price_lower_value))
  echo "$price.$((RANDOM % 99))"
}

get_slot_ids_from_db () {
  slot_ids=$(PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USERNAME} -d ${DB_DATABASE} -qt -c "select id from slot")
}

seed_screen_data () {
  echo "Seeding screen data..."

  screen_id=$(PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USERNAME} -d ${DB_DATABASE} -qtA -c \
  "insert into screen (screen_name) values ('Screen 1') returning id;")

  if [ -z "$screen_id" ]; then
    echo "ERROR: screen_id is empty!"
    exit 1
  fi

  echo "Screen created with ID: $screen_id"
}

initialise_dates () {
  if [ -n "$1" ]; then
    start_date="$1"
  else
    if [ "$OS_TYPE" = "Darwin" ]
    then
      start_date=$(date -j +"%Y-%m-%d")
    else
      start_date=$(date "+%Y-%m-%d")
    fi
  fi
  next_date="$start_date"
}

get_next_date () {
  # Try GNU date first (standard Linux)
  next_date=$(date -d "$1 + 1 day" "+%Y-%m-%d" 2>/dev/null)

  # If GNU date fails, try macOS/BSD date
  if [ -z "$next_date" ]; then
    next_date=$(date -j -f "%Y-%m-%d" -v+1d "$1" "+%Y-%m-%d" 2>/dev/null)
  fi

  # If both fail (Alpine/BusyBox), use epoch arithmetic
  if [ -z "$next_date" ]; then
    epoch=$(date -d "$1" +%s)
    next_date=$(date -d "@$((epoch + 86400))" "+%Y-%m-%d")
  fi

  echo "$next_date"
}

clear_old_data () {
  echo "Truncating the following tables in database: booking, show, slot, screen..."

  PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USERNAME} -d ${DB_DATABASE} -qc \
  "truncate booking, show, slot, screen cascade"

  echo "Tables successfully truncated!"
}

seed_data_for_first_day () {
  for slot_id in $slot_ids
  do
    movie_id=$(get_random_movie_id)
    price=$(get_random_price_with_two_decimal_places)
    PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USERNAME} -d ${DB_DATABASE} -qc \
    "insert into show (screen_id, movie_id, date, slot_id, cost) values ($screen_id, '$movie_id', '${start_date}', $slot_id, $price)"
  done
}

seed_data_second_day_onwards () {
  next_date=$(get_next_date "$start_date")
  # Check if number of days was provided as an argument
  total_days=${2:-21}  # Default to 21 days if not specified
  for day in $(seq 2 $total_days)
  do
    for slot_id in $slot_ids
    do
      movie_id=$(get_random_movie_id)
      price=$(get_random_price_with_two_decimal_places)
      PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USERNAME} -d ${DB_DATABASE} -qc \
      "insert into show (screen_id, movie_id, date, slot_id, cost) values ($screen_id, '$movie_id', '$next_date', $slot_id, $price)"
    done
    next_date=$(get_next_date "$next_date")
  done
}

seed_slot_data () {
  echo "Seeding slot data in database..."

  PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USERNAME} -d ${DB_DATABASE} -qc \
  "insert into slot (name, start_time, end_time) values \
  ('slot1', '09:00:00', '12:30:00'), \
  ('slot2', '13:30:00', '17:00:00'), \
  ('slot3', '18:00:00', '21:30:00'), \
  ('slot4', '22:30:00', '02:00:00')"

  echo "Slot data successfully seeded!"
}

seed_show_data () {
  # Use the provided number of days or default to 21
  total_days=${2:-21}
  echo "Seeding show data in database for $total_days days..."

  get_slot_ids_from_db
  seed_data_for_first_day
  seed_data_second_day_onwards

  echo "Show data successfully seeded!"
}

# Main execution
initialise_dates "$1"
clear_old_data
seed_screen_data
seed_slot_data
seed_show_data "$1" "$2"
