ALTER TABLE online_customers
DROP COLUMN IF EXISTS is_email_verified;

ALTER TABLE online_customers
DROP COLUMN IF EXISTS is_phone_verified;
