-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; 

CREATE TABLE IF NOT EXISTS online_customers (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL UNIQUE,
    phone_number VARCHAR(20),
    country_code TEXT NOT NULL, 
    age INT NOT NULL CHECK (age >= 0 AND age <= 130),
    gender VARCHAR(100) NOT NULL,
    username VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL,
    avatar_url TEXT,
    avatar_type VARCHAR(100),
    password_reset_token TEXT,
    password_token_expiry_mins INT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create the unique index for email as specified in your struct tags
CREATE UNIQUE INDEX IF NOT EXISTS idx_online_customers_email ON online_customers(email);