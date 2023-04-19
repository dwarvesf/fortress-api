
-- +migrate Up
ALTER TABLE accounting_transactions ADD COLUMN updated_at TIMESTAMP(6) DEFAULT (NOW());

-- +migrate Down
ALTER TABLE accounting_transactions DROP COLUMN updated_at;
