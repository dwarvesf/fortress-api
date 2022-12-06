
-- +migrate Up
ALTER TABLE project_members DROP CONSTRAINT IF EXISTS project_members_unique;

-- +migrate Down
ALTER TABLE project_members ADD CONSTRAINT project_members_unique UNIQUE (project_slot_id);
