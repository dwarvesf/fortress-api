-- +migrate Up
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS task_provider TEXT DEFAULT 'basecamp';
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS task_ref TEXT;
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS task_board TEXT;
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS task_attachment_url TEXT;

UPDATE expenses SET task_provider = 'basecamp' WHERE task_provider IS NULL;
UPDATE expenses SET task_ref = basecamp_id::text WHERE (task_ref IS NULL OR task_ref = '') AND basecamp_id IS NOT NULL;

ALTER TABLE expenses ALTER COLUMN task_provider SET NOT NULL;
ALTER TABLE expenses ALTER COLUMN task_provider DROP DEFAULT;

-- +migrate Down
ALTER TABLE expenses DROP COLUMN IF EXISTS task_attachment_url;
ALTER TABLE expenses DROP COLUMN IF EXISTS task_board;
ALTER TABLE expenses DROP COLUMN IF EXISTS task_ref;
ALTER TABLE expenses DROP COLUMN IF EXISTS task_provider;
