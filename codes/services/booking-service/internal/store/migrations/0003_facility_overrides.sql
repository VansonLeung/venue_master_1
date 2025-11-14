CREATE TABLE IF NOT EXISTS facility_overrides (
    id UUID PRIMARY KEY,
    facility_id UUID NOT NULL REFERENCES facilities(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    open_at TIME,
    close_at TIME,
    all_day BOOLEAN NOT NULL DEFAULT FALSE,
    reason TEXT,
    applies_weekdays INTEGER[] NOT NULL DEFAULT ARRAY[0,1,2,3,4,5,6],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_facility_overrides_dates
    ON facility_overrides (facility_id, start_date, end_date);
