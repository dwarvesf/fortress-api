
-- +migrate Up
ALTER TABLE project_slots DROP COLUMN IF EXISTS position;

ALTER TABLE project_members DROP COLUMN IF EXISTS position;

ALTER TABLE employees DROP CONSTRAINT employees_position_id_fkey;

ALTER TABLE employees DROP COLUMN position_id;

-- +migrate Down
ALTER TABLE employees ADD IF NOT EXISTS position_id UUID;

ALTER TABLE employees ADD CONSTRAINT employees_position_id_fkey FOREIGN KEY (position_id) REFERENCES positions (id);

ALTER TABLE project_members ADD IF NOT EXISTS position TEXT;

ALTER TABLE project_slots ADD IF NOT EXISTS position TEXT;
