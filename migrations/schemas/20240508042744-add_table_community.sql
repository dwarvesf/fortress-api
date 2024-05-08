-- +migrate Up
CREATE TABLE IF NOT EXISTS community_members (
    id UUID PRIMARY KEY DEFAULT ( UUID() ),
    discord_id  BIGINT,
    discord_username VARCHAR(40),
    role TEXT[],
    employee_id uuid NOT NULL,
    memo_username TEXT NOT NULL,
    github_username TEXT NOT NULL,
    personal_email TEXT NOT NULL
);
-- +migrate Down
DROP TABLE IF EXISTS community_members;
