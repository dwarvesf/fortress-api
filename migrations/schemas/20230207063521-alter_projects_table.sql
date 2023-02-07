-- +migrate Up
ALTER TABLE projects ADD COLUMN "bank_account_id" UUID;
ALTER TABLE projects ADD CONSTRAINT projects_bank_account_id_fkey FOREIGN KEY (bank_account_id) REFERENCES bank_accounts (id);

-- +migrate Down
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_bank_account_id_fkey;
ALTER TABLE projects DROP COLUMN IF EXISTS "bank_account_id";
