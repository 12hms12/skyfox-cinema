DROP INDEX IF EXISTS idx_show_screen_date_slot;

ALTER TABLE "show"
ADD CONSTRAINT show_date_slot_id_key UNIQUE (date, slot_id);