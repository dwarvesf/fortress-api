-- +migrate Up
ALTER TABLE employees ADD COLUMN "referred_by" UUID;
ALTER TABLE employees ADD CONSTRAINT employees_referred_by_fkey FOREIGN KEY (referred_by) REFERENCES employees (id);

-- +migrate Down
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_referred_by_fkey;
ALTER TABLE employees DROP COLUMN IF EXISTS "referred_by";

