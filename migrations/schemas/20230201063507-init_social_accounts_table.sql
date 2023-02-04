
-- +migrate Up
CREATE TABLE IF NOT EXISTS social_accounts (
    id              uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at      TIMESTAMP(6),
    created_at      TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at      TIMESTAMP(6)     DEFAULT (NOW()),
    employee_id     uuid,
    type            TEXT,
    account_id      TEXT, -- e.g. id use for integrate with social platform.
    name            TEXT,
    email           TEXT
);

ALTER TABLE social_accounts
    ADD CONSTRAINT social_accounts_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);

-- +migrate Down
ALTER TABLE social_accounts DROP CONSTRAINT IF EXISTS social_accounts_employee_id_fkey;

DROP TABLE IF EXISTS social_accounts;
