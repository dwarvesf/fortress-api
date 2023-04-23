
-- +migrate Up
ALTER TABLE base_salaries ADD COLUMN updated_at TIMESTAMP(6) DEFAULT (NOW());

-- +migrate Down
ALTER TABLE base_salaries DROP COLUMN updated_at;
