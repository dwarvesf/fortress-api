-- +migrate Up
ALTER TABLE employees ALTER  COLUMN  team_email DROP NOT NULL;

-- +migrate Down
