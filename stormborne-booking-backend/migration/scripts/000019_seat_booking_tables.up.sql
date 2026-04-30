CREATE TABLE booking (
    id SERIAL PRIMARY KEY,
    transaction_id VARCHAR(255),

    online_customer_id INTEGER,
    show_id INTEGER,

    total_price NUMERIC(10,2),
    status VARCHAR(20),

    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_booking_customer
        FOREIGN KEY (online_customer_id)
        REFERENCES online_customers(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_booking_show
        FOREIGN KEY (show_id)
        REFERENCES show(id)
        ON DELETE CASCADE
);


CREATE TABLE show_seat_status (
    id SERIAL PRIMARY KEY,

    show_id INTEGER,
    seat_id INTEGER,

    status VARCHAR(20),
    locked_until TIMESTAMP NULL,

    CONSTRAINT fk_showseatstatus_show
        FOREIGN KEY (show_id)
        REFERENCES show(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_showseatstatus_seat
        FOREIGN KEY (seat_id)
        REFERENCES seat(id)
        ON DELETE CASCADE,

    CONSTRAINT unique_show_seat UNIQUE (show_id, seat_id)
);


CREATE TABLE show_pricing (
    id SERIAL PRIMARY KEY,

    show_id INTEGER,
    seat_type VARCHAR(50),
    price NUMERIC(10,2),

    CONSTRAINT fk_showpricing_show
        FOREIGN KEY (show_id)
        REFERENCES show(id)
        ON DELETE CASCADE
);


CREATE TABLE booked_seat (
    id SERIAL PRIMARY KEY,

    booking_id INTEGER,
    show_seat_status_id INTEGER,

    CONSTRAINT fk_bookedseat_booking
        FOREIGN KEY (booking_id)
        REFERENCES booking(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_bookedseat_status
        FOREIGN KEY (show_seat_status_id)
        REFERENCES show_seat_status(id)
        ON DELETE CASCADE
);


-- Indexes
CREATE INDEX idx_booking_show_id ON booking(show_id);
CREATE INDEX idx_booking_customer_id ON booking(online_customer_id);

CREATE INDEX idx_show_seat_status_show_id ON show_seat_status(show_id);
CREATE INDEX idx_show_seat_status_seat_id ON show_seat_status(seat_id);

CREATE INDEX idx_booked_seat_booking_id ON booked_seat(booking_id);
CREATE INDEX idx_booked_seat_status_id ON booked_seat(show_seat_status_id);