ALTER TABLE "show"
DROP CONSTRAINT IF EXISTS show_date_slot_id_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_show_screen_date_slot
ON "show"(screen_id, date, slot_id);