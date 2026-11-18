CREATE TABLE IF NOT EXISTS leases (
    id         BIGSERIAL PRIMARY KEY,
    tenant_id  BIGINT      NOT NULL REFERENCES users (id),
    room_id    BIGINT      NOT NULL REFERENCES rooms (id),
    start_date DATE        NOT NULL,
    end_date   DATE        NOT NULL,
    status     VARCHAR(10) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'ended')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_date > start_date)
);

CREATE UNIQUE INDEX IF NOT EXISTS leases_one_active_per_room
    ON leases (room_id) WHERE status = 'active';

CREATE INDEX IF NOT EXISTS leases_tenant_idx ON leases (tenant_id);
