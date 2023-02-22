
-- +migrate Up
ALTER TABLE project_slots ADD COLUMN note TEXT;
ALTER TABLE project_members ADD COLUMN note TEXT;

-- +migrate Down
ALTER TABLE project_members DROP COLUMN IF EXISTS note;
ALTER TABLE project_slots DROP COLUMN IF EXISTS note;
