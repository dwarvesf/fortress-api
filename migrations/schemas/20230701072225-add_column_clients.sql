-- +migrate Up
ALTER TABLE clients ADD COLUMN city TEXT DEFAULT NULL;
ALTER TABLE clients ADD COLUMN company_size INT DEFAULT 0;
ALTER TABLE clients ADD COLUMN solution_type TEXT DEFAULT NULL;
ALTER TABLE clients ADD COLUMN lat TEXT DEFAULT NULL;
ALTER TABLE clients ADD COLUMN long TEXT DEFAULT NULL;
ALTER TABLE clients ADD COLUMN is_public BOOLEAN DEFAULT FALSE;

-- +migrate Down
ALTER TABLE clients DROP COLUMN city;
ALTER TABLE clients DROP COLUMN company_size;
ALTER TABLE clients DROP COLUMN solution_type;
ALTER TABLE clients DROP COLUMN lat;
ALTER TABLE clients DROP COLUMN long;
ALTER TABLE clients DROP COLUMN is_public;