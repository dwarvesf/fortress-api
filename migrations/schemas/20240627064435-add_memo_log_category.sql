
-- +migrate Up
ALTER TABLE memo_logs ADD COLUMN "category" TEXT[] default '{}'::TEXT[];

-- +migrate Down
ALTER TABLE memo_logs DROP COLUMN "category";
