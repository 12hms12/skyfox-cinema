ALTER TABLE booking
DROP CONSTRAINT IF EXISTS fk_booking_show;

ALTER TABLE booking
ADD CONSTRAINT fk_booking_show
FOREIGN KEY (show_id)
REFERENCES "show"(id)
ON DELETE CASCADE;

ALTER TABLE show_seat_status
DROP CONSTRAINT IF EXISTS fk_showseatstatus_show;

ALTER TABLE show_seat_status
ADD CONSTRAINT fk_showseatstatus_show
FOREIGN KEY (show_id)
REFERENCES "show"(id)
ON DELETE CASCADE;

ALTER TABLE show_pricing
DROP CONSTRAINT IF EXISTS fk_showpricing_show;

ALTER TABLE show_pricing
ADD CONSTRAINT fk_showpricing_show
FOREIGN KEY (show_id)
REFERENCES "show"(id)
ON DELETE CASCADE;
