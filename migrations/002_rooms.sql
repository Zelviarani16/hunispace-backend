CREATE TABLE IF NOT EXISTS rooms (
    id            BIGSERIAL PRIMARY KEY,
    room_number   VARCHAR(20)   NOT NULL UNIQUE,
    room_type     VARCHAR(50)   NOT NULL,
    monthly_price NUMERIC(12,2) NOT NULL CHECK (monthly_price > 0),
    facilities    TEXT          NOT NULL DEFAULT '',
    status        VARCHAR(12)   NOT NULL DEFAULT 'available'
                  CHECK (status IN ('available', 'occupied', 'maintenance')),
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT now()
);
