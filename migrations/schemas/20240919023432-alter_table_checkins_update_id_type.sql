
-- +migrate Up
ALTER TABLE physical_checkins DROP CONSTRAINT physical_checkins_pkey;
ALTER TABLE physical_checkins DROP COLUMN id;
ALTER TABLE physical_checkins ADD COLUMN id uuid PRIMARY KEY DEFAULT (uuid());
ALTER TABLE discord_checkins DROP CONSTRAINT discord_checkins_pkey;
ALTER TABLE discord_checkins DROP COLUMN id;
ALTER TABLE discord_checkins ADD COLUMN id uuid PRIMARY KEY DEFAULT (uuid());

-- +migrate Down
ALTER TABLE physical_checkins DROP CONSTRAINT physical_checkins_pkey;
ALTER TABLE physical_checkins ADD COLUMN id SERIAL PRIMARY KEY;
ALTER TABLE discord_checkins DROP CONSTRAINT discord_checkins_pkey;
ALTER TABLE discord_checkins ADD COLUMN id SERIAL PRIMARY KEY;