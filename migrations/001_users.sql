CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    full_name     VARCHAR(100) NOT NULL,
    email         VARCHAR(150) NOT NULL UNIQUE,
    password_hash VARCHAR(100) NOT NULL,
    phone_number  VARCHAR(20)  NOT NULL DEFAULT '',
    role          VARCHAR(10)  NOT NULL DEFAULT 'tenant' CHECK (role IN ('owner', 'tenant')),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);
