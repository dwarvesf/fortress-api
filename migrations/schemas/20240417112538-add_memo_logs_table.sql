-- +migrate Up
CREATE TABLE IF NOT EXISTS memo_logs (
    id              UUID PRIMARY KEY DEFAULT (UUID()),
    deleted_at      TIMESTAMP(6),
    created_at      TIMESTAMP(6) DEFAULT (now()),
    updated_at      TIMESTAMP(6) DEFAULT (now()),

    title           TEXT NOT NULL,
    url             TEXT NOT NULL,
    authors         JSONB,
    tags            JSONB,
    description     TEXT,
    published_at    TIMESTAMP(6) NOT NULL,
    reward          DECIMAL
);

-- +migrate Down
DROP TABLE IF EXISTS memo_logs;
