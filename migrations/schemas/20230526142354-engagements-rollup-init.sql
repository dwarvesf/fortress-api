-- +migrate Up
CREATE TABLE engagements_rollup
(
    id                uuid PRIMARY KEY DEFAULT (uuid()),
    discord_user_id   TEXT,
    latest_message_id TEXT,
    discord_username  VARCHAR(40),
    -- Discord's usernames have the form "username#number"
    -- for example: thanhnguyen2187#4183
    -- the username part's maximal length is 32, while the number part's length is 4
    channel_id        TEXT,
    category_id       TEXT,
    message_count     INT,
    reaction_count    INT,
    deleted_at        TIMESTAMP(6),
    created_at        TIMESTAMP(6),
    updated_at        TIMESTAMP(6),

    UNIQUE (discord_user_id, channel_id)
);
-- using TEXT instead of BIGINT for discord_user_id
-- and for                          latest_message_id
-- and for                          channel_id
-- and for                          category_id
-- since the type for those numbers uint64,
-- while Postgres's BIGINT is int64
-- https://www.postgresql.org/docs/current/datatype-numeric.html
-- https://discord.com/developers/docs/reference#snowflakes

-- +migrate Down
DROP TABLE engagements_rollup;