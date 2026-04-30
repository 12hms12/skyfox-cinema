CREATE TABLE IF NOT EXISTS seat (
    id SERIAL PRIMARY KEY,
    screen_id INT NOT NULL,
    row_number INT,
    column_number INT,
    seat_type VARCHAR(50),

    CONSTRAINT fk_seat_screen
        FOREIGN KEY (screen_id)
        REFERENCES screen(id)
        ON DELETE CASCADE
);