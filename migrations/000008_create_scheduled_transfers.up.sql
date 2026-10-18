CREATE TABLE scheduled_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_username VARCHAR(255) NOT NULL,
    amount NUMERIC(19,4) NOT NULL CHECK (amount > 0),
    frequency VARCHAR(20) NOT NULL CHECK (frequency IN ('daily','weekly','monthly')),
    next_run_at TIMESTAMPTZ NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    last_run_at TIMESTAMPTZ,
    last_run_status VARCHAR(20),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_scheduled_transfers_next_run ON scheduled_transfers(next_run_at) WHERE active = true;
CREATE INDEX idx_scheduled_transfers_user ON scheduled_transfers(user_id);
