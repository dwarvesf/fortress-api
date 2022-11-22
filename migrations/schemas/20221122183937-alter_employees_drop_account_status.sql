
-- +migrate Up
ALTER TABLE employees DROP COLUMN IF EXISTS account_status;

-- +migrate Down
ALTER TABLE employees ADD IF NOT EXISTS account_status working_status;
