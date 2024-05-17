-- +migrate Up
ALTER TABLE discord_accounts
RENAME COLUMN username TO discord_username;

UPDATE discord_accounts 
SET discord_username = '' WHERE discord_username IS NULL;

ALTER TABLE discord_accounts  
ALTER COLUMN discord_username SET NOT NULL,
ALTER COLUMN discord_username SET DEFAULT '';

ALTER TABLE discord_accounts
ADD COLUMN roles TEXT[] DEFAULT NULL,
ADD COLUMN memo_username TEXT NOT NULL DEFAULT '',
ADD COLUMN github_username TEXT NOT NULL DEFAULT '',
ADD COLUMN personal_email TEXT NOT NULL DEFAULT '';

-- +migrate Down
ALTER TABLE discord_accounts
DROP COLUMN roles,
DROP COLUMN memo_username,
DROP COLUMN github_username,
DROP COLUMN personal_email;

ALTER TABLE discord_accounts
ALTER COLUMN discord_username DROP DEFAULT,
ALTER COLUMN discord_username DROP NOT NULL;

UPDATE discord_accounts
SET discord_username = NULL WHERE discord_username = '';

ALTER TABLE discord_accounts
RENAME COLUMN discord_username TO username;
