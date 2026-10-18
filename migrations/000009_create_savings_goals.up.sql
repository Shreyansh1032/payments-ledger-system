CREATE TABLE savings_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    target_amount NUMERIC(19,4) NOT NULL CHECK (target_amount > 0),
    current_amount NUMERIC(19,4) NOT NULL DEFAULT 0 CHECK (current_amount >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_savings_goals_user ON savings_goals(user_id);
