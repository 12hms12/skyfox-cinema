ALTER TABLE online_customers 
DROP COLUMN IF EXISTS password_token_expiry_mins;

ALTER TABLE online_customers 
ADD COLUMN IF NOT EXISTS password_token_expiry_time TIMESTAMP WITH TIME ZONE;
