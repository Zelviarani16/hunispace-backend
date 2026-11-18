CREATE TABLE IF NOT EXISTS maintenance_tickets (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   BIGINT       NOT NULL REFERENCES users (id),
    room_id     BIGINT       NOT NULL REFERENCES rooms (id),
    title       VARCHAR(150) NOT NULL,
    description TEXT         NOT NULL,
    priority    VARCHAR(10)  NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
    status      VARCHAR(12)  NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'resolved')),
    version     INT          NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS maintenance_tickets_tenant_idx ON maintenance_tickets (tenant_id);
