
-- +migrate Up
ALTER TABLE projects ADD COLUMN "organization_id" UUID;
ALTER TABLE projects
    ADD CONSTRAINT projects_organizations_id_fkey FOREIGN KEY (organization_id) REFERENCES organizations (id);

-- +migrate Down
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_organizations_id_fkey;
ALTER TABLE projects DROP COLUMN IF EXISTS "organization_id";
