-- +migrate Down
ALTER TABLE usertable
DROP COLUMN name,
DROP COLUMN counter_no;