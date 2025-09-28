-- +goose Up
-- Create contacts table for amateur radio QSO logging
CREATE TABLE contacts (
    id SERIAL PRIMARY KEY,
    callsign VARCHAR(20) NOT NULL,
    contact_date DATE NOT NULL,
    time_on VARCHAR(4),
    time_off VARCHAR(4),
    frequency DECIMAL(10, 6),
    band VARCHAR(10),
    mode VARCHAR(10),
    rst_sent VARCHAR(10),
    rst_received VARCHAR(10),
    operator_name VARCHAR(100),
    qth VARCHAR(100),
    country VARCHAR(50),
    grid_square VARCHAR(10),
    power_watts INTEGER,
    comment TEXT,
    confirmed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for common queries
CREATE INDEX idx_contacts_callsign ON contacts(callsign);
CREATE INDEX idx_contacts_date ON contacts(contact_date);
CREATE INDEX idx_contacts_band ON contacts(band);
CREATE INDEX idx_contacts_mode ON contacts(mode);
CREATE INDEX idx_contacts_country ON contacts(country);

-- +goose Down
-- Drop indexes
DROP INDEX IF EXISTS idx_contacts_country;
DROP INDEX IF EXISTS idx_contacts_mode;
DROP INDEX IF EXISTS idx_contacts_band;
DROP INDEX IF EXISTS idx_contacts_date;
DROP INDEX IF EXISTS idx_contacts_callsign;

-- Drop table
DROP TABLE IF EXISTS contacts;