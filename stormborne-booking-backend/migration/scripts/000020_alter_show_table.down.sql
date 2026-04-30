DROP INDEX IF EXISTS idx_show_screen_id;

ALTER TABLE "show"
DROP CONSTRAINT IF EXISTS fk_show_screen;

ALTER TABLE "show"
DROP COLUMN IF EXISTS screen_id;