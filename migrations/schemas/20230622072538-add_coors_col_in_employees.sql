-- +migrate Up
ALTER TABLE employees ADD COLUMN lat TEXT DEFAULT NULL;
ALTER TABLE employees ADD COLUMN long TEXT DEFAULT NULL;
-- +migrate Down
ALTER TABLE employees DROP COLUMN lat;
ALTER TABLE employees DROP COLUMN long;
