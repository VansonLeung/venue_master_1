ALTER TABLE facilities
    ADD COLUMN IF NOT EXISTS weekday_rate_cents INTEGER NOT NULL DEFAULT 4500,
    ADD COLUMN IF NOT EXISTS weekend_rate_cents INTEGER NOT NULL DEFAULT 6000,
    ADD COLUMN IF NOT EXISTS currency TEXT NOT NULL DEFAULT 'CAD';

CREATE TABLE IF NOT EXISTS payment_retries (
    booking_id UUID PRIMARY KEY REFERENCES bookings(id) ON DELETE CASCADE,
    attempt INTEGER NOT NULL,
    next_attempt_at TIMESTAMPTZ NOT NULL,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_retries_next ON payment_retries(next_attempt_at);
