-- +migrate Up
-- Create physical_checkins table
CREATE TABLE physical_checkins (
    id SERIAL PRIMARY KEY,
    employee_id UUID REFERENCES employees(id),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    date DATE
);

-- Create discord_checkins table
CREATE TABLE discord_checkins (
    id SERIAL PRIMARY KEY,
    employee_id UUID REFERENCES employees(id),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    date DATE
);

-- Create index on date
CREATE INDEX idx_physical_checkins_date ON physical_checkins (date);
CREATE INDEX idx_discord_checkins_date ON discord_checkins (date);

-- +migrate Down
-- Drop physical_checkins table
DROP TABLE physical_checkins;
-- Drop discord_checkins table
DROP TABLE discord_checkins;
