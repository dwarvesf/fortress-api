
-- +migrate Up
ALTER TYPE employment_status RENAME TO working_status;
ALTER TABLE employees RENAME COLUMN employment_status TO "working_status";
-- +migrate Down
ALTER TYPE working_status rename TO employment_status;
ALTER TABLE employees RENAME COLUMN working_status TO "employment_status";
