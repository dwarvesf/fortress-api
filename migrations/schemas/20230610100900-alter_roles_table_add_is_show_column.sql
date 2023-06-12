-- +migrate Up
ALTER TABLE roles ADD COLUMN is_show BOOL NOT NULL DEFAULT TRUE;
ALTER TABLE roles ADD COLUMN color TEXT;
-- +migrate Down
ALTER TABLE roles DROP COLUMN is_show;
ALTER TABLE roles DROP COLUMN color;
