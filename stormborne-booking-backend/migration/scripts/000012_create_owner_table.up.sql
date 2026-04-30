CREATE TABLE IF NOT EXISTS owner (
    id SERIAL PRIMARY KEY,
    name TEXT,
    username TEXT UNIQUE,
    password TEXT
);