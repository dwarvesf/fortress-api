
-- +migrate Up
ALTER TABLE "projects" ADD COLUMN "code" TEXT;
ALTER TABLE projects ADD UNIQUE (code);

-- +migrate Down
ALTER TABLE "projects" DROP CONSTRAINT IF EXISTS projects_code_key;
ALTER TABLE "projects" DROP COLUMN "code";
