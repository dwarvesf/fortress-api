
-- +migrate Up
ALTER TABLE memo_logs DROP COLUMN IF EXISTS authors;
-- +migrate Down
ALTER TABLE memo_logs ADD COLUMN authors JSONB;
