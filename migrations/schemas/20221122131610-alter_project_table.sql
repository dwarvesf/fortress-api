
-- +migrate Up
ALTER TABLE projects DROP COLUMN IF EXISTS country;
ALTER TABLE projects ADD IF NOT EXISTS country_id UUID;
ALTER TABLE projects ADD IF NOT EXISTS client_email TEXT;
ALTER TABLE projects ADD IF NOT EXISTS project_email TEXT;

-- +migrate Down
ALTER TABLE projects DROP COLUMN IF EXISTS country_id;
ALTER TABLE projects DROP COLUMN IF EXISTS client_email;
ALTER TABLE projects DROP COLUMN IF EXISTS project_email;
ALTER TABLE projects ADD IF NOT EXISTS country Text;
