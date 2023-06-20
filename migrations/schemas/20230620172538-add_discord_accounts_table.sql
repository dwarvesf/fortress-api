-- +migrate Up
CREATE TABLE IF NOT EXISTS discord_accounts (
    id         UUID PRIMARY KEY DEFAULT UUID(),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT NOW(),
    updated_at TIMESTAMP(6)     DEFAULT NOW(),

    discord_id TEXT NOT NULL,
    username   TEXT NOT NULL
);


ALTER TABLE employees ADD COLUMN discord_account_id UUID DEFAULT NULL;
ALTER TABLE employees
    ADD CONSTRAINT employees_discord_account_id_fkey FOREIGN KEY (discord_account_id) REFERENCES discord_accounts (id);

ALTER TABLE discord_accounts ADD UNIQUE (discord_id);
-- +migrate Down
ALTER TABLE employees DROP COLUMN discord_account_id;

DROP TABLE IF EXISTS discord_accounts;
