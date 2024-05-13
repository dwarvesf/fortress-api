-- +migrate Up
CREATE TABLE IF NOT EXISTS community_members (
    id UUID PRIMARY KEY DEFAULT UUID(),
    discord_id TEXT NOT NULL UNIQUE,
    discord_username VARCHAR(40) NOT NULL DEFAULT '' UNIQUE,
    roles TEXT[],
    employee_id UUID REFERENCES employees(id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION
        DEFAULT NULL,
    memo_username TEXT NOT NULL DEFAULT '',
    github_username TEXT NOT NULL DEFAULT '',
    personal_email TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMP(6)     DEFAULT (now()),
    updated_at    TIMESTAMP(6)     DEFAULT (now()),
    deleted_at    TIMESTAMP(6)
);

-- +migrate Down
DROP TABLE IF EXISTS community_members;
