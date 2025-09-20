-- +goose Up
-- Update time_on and time_off columns to support HH:MM:SS format
ALTER TABLE contacts 
    ALTER COLUMN time_on TYPE VARCHAR(8),
    ALTER COLUMN time_off TYPE VARCHAR(8);

-- +goose Down
-- Revert time_on and time_off columns back to VARCHAR(4)
ALTER TABLE contacts 
    ALTER COLUMN time_on TYPE VARCHAR(4),
    ALTER COLUMN time_off TYPE VARCHAR(4);