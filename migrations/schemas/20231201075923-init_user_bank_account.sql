-- +migrate Up
CREATE TABLE IF NOT EXISTS "banks" (
    id         UUID PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at TIMESTAMP(6)     DEFAULT (NOW()),

    "name"       TEXT DEFAULT NULL,
    "code"       TEXT DEFAULT NULL,
    "bin"        TEXT DEFAULT NULL,
    "short_name" TEXT DEFAULT NULL,
    "logo"       TEXT DEFAULT NULL,
    "swift_code" TEXT DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS "user_bank_accounts" (
    id         UUID PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at TIMESTAMP(6)     DEFAULT (NOW()),

    "discord_account_id" UUID,
    "employee_id"        UUID NULL,
    "bank_id"            UUID,
    "account_number"     TEXT DEFAULT NULL,
    "branch"             TEXT DEFAULT NULL
);

ALTER TABLE "user_bank_accounts"
    ADD FOREIGN KEY ("discord_account_id") REFERENCES "discord_accounts" ("id");

ALTER TABLE "user_bank_accounts"
    ADD FOREIGN KEY ("employee_id") REFERENCES "employees" ("id");
-- +migrate Down
DROP TABLE IF EXISTS "user_bank_accounts";
DROP TABLE IF EXISTS "banks";
