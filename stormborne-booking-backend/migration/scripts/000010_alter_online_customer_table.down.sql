-- Remove email unique constraint
ALTER TABLE online_customers
DROP CONSTRAINT IF EXISTS uq_online_customers_email;

-- Recreate old index
CREATE UNIQUE INDEX idx_online_customers_email ON online_customers(email);

-- Remove phone constraint
ALTER TABLE online_customers
DROP CONSTRAINT IF EXISTS uq_online_customers_phone;

-- Remove is_verified column
ALTER TABLE online_customers
DROP COLUMN IF EXISTS is_verified;

