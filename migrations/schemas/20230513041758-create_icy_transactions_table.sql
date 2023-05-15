-- +migrate Up
CREATE TABLE IF NOT EXISTS icy_transactions (
    id                   uuid PRIMARY KEY     DEFAULT (uuid()),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMPTZ,
    vault                TEXT        NOT NULL,
    amount               TEXT        NOT NULL,
    token                TEXT        NOT NULL,
    sender_discord_id    TEXT        NOT NULL,
    recipient_address    TEXT        NOT NULL,
    recipient_discord_id TEXT        NOT NULL
);
-- +migrate Down
DROP TABLE IF EXISTS icy_transactions;
