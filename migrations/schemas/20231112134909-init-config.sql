-- +migrate Up
CREATE TABLE IF NOT EXISTS configs (
    id          uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at  TIMESTAMP(6),
    created_at  TIMESTAMP(6)     DEFAULT (now()),
    updated_at  TIMESTAMP(6)     DEFAULT (now()),
    key        TEXT NOT NULL,
    value      TEXT NOT NULL
);

ALTER TABLE configs ADD UNIQUE (key);

-- +migrate Down
DROP TABLE IF EXISTS configs;
