ALTER TABLE "show"
ADD COLUMN screen_id INTEGER;

ALTER TABLE "show"
ADD CONSTRAINT fk_show_screen
FOREIGN KEY (screen_id)
REFERENCES screen(id)
ON DELETE CASCADE;

CREATE INDEX idx_show_screen_id
ON "show"(screen_id);