-- +migrate Up
ALTER TABLE employees ADD IF NOT EXISTS notion_id TEXT;

-- +migrate Down
ALTER TABLE employees DROP COLUMN IF EXISTS notion_id;
