CREATE TABLE IF NOT EXISTS invoices (
    id             BIGSERIAL PRIMARY KEY,
    lease_id       BIGINT        NOT NULL REFERENCES leases (id),
    invoice_number VARCHAR(40)   NOT NULL UNIQUE,
    billing_period VARCHAR(7)    NOT NULL,
    amount         NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    due_date       DATE          NOT NULL,
    payment_method VARCHAR(30),
    proof_url      TEXT,
    status         VARCHAR(22)   NOT NULL DEFAULT 'unpaid'
                   CHECK (status IN ('unpaid', 'pending_verification', 'paid', 'overdue')),
    paid_at        TIMESTAMPTZ,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),

    UNIQUE (lease_id, billing_period)
);
