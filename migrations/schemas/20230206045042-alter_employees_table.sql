-- +migrate Up
ALTER TABLE employees ADD COLUMN "referred_by" UUID;
ALTER TABLE employees ADD CONSTRAINT employees_referred_by_fkey FOREIGN KEY (referred_by) REFERENCES employees (id);
ALTER TABLE projects ADD COLUMN "bank_account_id" UUID;
ALTER TABLE projects ADD CONSTRAINT projects_bank_account_id_fkey FOREIGN KEY (bank_account_id) REFERENCES bank_accounts (id);

-- +migrate Down
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_referred_by_fkey;
ALTER TABLE employees DROP COLUMN IF EXISTS "referred_by";
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_bank_account_id_fkey;
ALTER TABLE projects DROP COLUMN IF EXISTS "bank_account_id";
