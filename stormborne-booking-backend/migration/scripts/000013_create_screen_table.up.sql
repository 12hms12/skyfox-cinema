CREATE TABLE IF NOT EXISTS screen (
    id SERIAL PRIMARY KEY,
    screen_name TEXT,
    owner_id INT,
    CONSTRAINT fk_screen_owner
        FOREIGN KEY (owner_id)
        REFERENCES owner(id)
        ON DELETE CASCADE
);