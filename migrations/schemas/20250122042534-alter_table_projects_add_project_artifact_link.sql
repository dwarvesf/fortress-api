
-- +migrate Up
ALTER TABLE projects ADD COLUMN artifact_link VARCHAR(255);

-- +migrate Down
ALTER TABLE projects DROP COLUMN artifact_link;
