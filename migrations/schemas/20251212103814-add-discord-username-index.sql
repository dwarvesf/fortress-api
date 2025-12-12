-- +migrate Up
CREATE INDEX IF NOT EXISTS idx_discord_accounts_username
ON discord_accounts(discord_username);

-- +migrate Down
DROP INDEX IF EXISTS idx_discord_accounts_username;
