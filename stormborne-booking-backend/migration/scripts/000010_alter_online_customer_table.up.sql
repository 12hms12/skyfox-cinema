-- Remove old unique index created earlier
DROP INDEX IF EXISTS idx_online_customers_email;

-- Add unique constraint for email (only if it doesn't exist)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_online_customers_email'
    ) THEN
        ALTER TABLE online_customers
        ADD CONSTRAINT uq_online_customers_email UNIQUE (email);
    END IF;
END $$;

-- Make phone_number NOT NULL (only if currently nullable)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'online_customers'
        AND column_name = 'phone_number'
        AND is_nullable = 'YES'
    ) THEN
        ALTER TABLE online_customers
        ALTER COLUMN phone_number SET NOT NULL;
    END IF;
END $$;

-- Add unique constraint for phone_number (only if it doesn't exist)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'uq_online_customers_phone'
    ) THEN
        ALTER TABLE online_customers
        ADD CONSTRAINT uq_online_customers_phone UNIQUE (phone_number);
    END IF;
END $$;

-- Add is_verified column
ALTER TABLE online_customers
ADD COLUMN IF NOT EXISTS is_verified BOOLEAN DEFAULT FALSE;