CREATE UNIQUE INDEX IF NOT EXISTS idx_show_screen_date_slot
ON "show"(screen_id, date, slot_id);