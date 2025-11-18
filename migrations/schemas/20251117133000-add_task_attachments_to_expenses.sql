-- +migrate Up
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS task_attachments JSONB DEFAULT '[]'::jsonb;
UPDATE expenses SET task_attachments = '[]'::jsonb WHERE task_attachments IS NULL;
ALTER TABLE expenses ALTER COLUMN task_attachments SET NOT NULL;

-- +migrate Down
ALTER TABLE expenses DROP COLUMN IF EXISTS task_attachments;
