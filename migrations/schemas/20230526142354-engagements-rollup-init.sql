-- +migrate Up
CREATE TABLE engagements_rollup
(
    id               uuid PRIMARY KEY DEFAULT (uuid()),
    discord_user_id  BIGINT,
    last_message_id  BIGINT,
    discord_username VARCHAR(40),
    -- Discord's usernames have the form "username#number"
    -- for example: thanhnguyen2187#4183
    -- the username part's maximal length is 32, while the number part's length is 4
    channel_id       BIGINT,
    category_id      BIGINT,
    message_count    INT NOT NULL DEFAULT 0,
    reaction_count   INT NOT NULL DEFAULT 0,
    deleted_at       TIMESTAMP(6),
    created_at       TIMESTAMP(6),
    updated_at       TIMESTAMP(6),

    UNIQUE (discord_user_id, channel_id)
);
-- TODO: migrate before 2084, since BIGINT of Postgres is int64
--       its maximal value is 9223372036854775807, which is equivalent to 2084-09-06T15:47:35.551Z
-- https://www.postgresql.org/docs/current/datatype-numeric.html
-- https://discord.com/developers/docs/reference#snowflakes
-- https://snowsta.mp/?l=en&z=n&f=ixetipsqxx-c73

-- +migrate Down
DROP TABLE engagements_rollup;

