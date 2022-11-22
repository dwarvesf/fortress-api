
-- +migrate Up
ALTER TABLE project_heads DROP CONSTRAINT IF EXISTS project_heads_unique;

-- +migrate Down
ALTER TABLE project_heads ADD CONSTRAINT project_heads_unique UNIQUE (project_id, employee_id, position);
