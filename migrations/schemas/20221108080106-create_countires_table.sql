-- +migrate Up
CREATE TABLE IF NOT EXISTS countries (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (now()),
    updated_at TIMESTAMP(6)     DEFAULT (now()),
    name       TEXT,
    code       TEXT,
    cities     JSONB            DEFAULT '[]'::JSONB
);

-- +migrate Down
DROP TABLE IF EXISTS countries;
