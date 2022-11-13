
-- +migrate Up
ALTER TABLE projects ADD IF NOT EXISTS country TEXT;

-- +migrate Down
ALTER TABLE projects DROP COLUMN IF EXISTS country;
