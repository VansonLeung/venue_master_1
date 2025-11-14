-- Create venues table
CREATE TABLE IF NOT EXISTS venues (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    description TEXT,
    address TEXT,
    city TEXT,
    state TEXT,
    zip_code TEXT,
    country TEXT DEFAULT 'US',
    phone TEXT,
    email TEXT,
    website TEXT,
    timezone TEXT DEFAULT 'America/New_York',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add trigger for venues updated_at
DROP TRIGGER IF EXISTS venues_set_updated_at ON venues;
CREATE TRIGGER venues_set_updated_at
BEFORE UPDATE ON venues
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- Insert default venue FIRST (before adding foreign key constraint)
INSERT INTO venues (id, name, description, address, city, state, country, timezone)
VALUES (
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    'Main Venue',
    'Primary venue location',
    '123 Main Street',
    'New York',
    'NY',
    'US',
    'America/New_York'
)
ON CONFLICT (id) DO NOTHING;

-- Now add foreign key constraint to facilities table
ALTER TABLE facilities
    DROP CONSTRAINT IF EXISTS fk_facilities_venue_id;

ALTER TABLE facilities
    ADD CONSTRAINT fk_facilities_venue_id
    FOREIGN KEY (venue_id) REFERENCES venues(id) ON DELETE CASCADE;

-- Create index on venue_id
CREATE INDEX IF NOT EXISTS idx_facilities_venue_id ON facilities(venue_id);
