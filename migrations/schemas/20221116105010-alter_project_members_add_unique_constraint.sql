
-- +migrate Up
ALTER TABLE project_heads ADD CONSTRAINT project_heads_unique UNIQUE (project_id, employee_id, position);

ALTER TABLE project_members ADD CONSTRAINT project_members_unique UNIQUE (project_slot_id);

ALTER TABLE project_member_positions ADD CONSTRAINT project_member_positions_unique UNIQUE (project_member_id, position_id);

ALTER TABLE project_members DROP COLUMN IF EXISTS position;

-- +migrate Down
ALTER TABLE project_heads DROP CONSTRAINT project_heads_unique;

ALTER TABLE project_members DROP CONSTRAINT project_members_unique;

ALTER TABLE project_member_positions DROP CONSTRAINT project_member_positions_unique;

ALTER TABLE project_members ADD IF NOT EXISTS position TEXT;
