-- +migrate Up
DELETE FROM action_item_snapshots;
DELETE FROM audit_participants;
DELETE FROM audit_action_items;
DELETE FROM action_items;
DELETE FROM audit_items;
DELETE FROM audits;
DELETE FROM audit_cycles;
ALTER TABLE audit_cycles DROP CONSTRAINT IF EXISTS audit_cycles_project_notion_id_fkey;
ALTER TABLE audits DROP CONSTRAINT IF EXISTS audits_project_notion_id_fkey;
ALTER TABLE action_items DROP CONSTRAINT IF EXISTS action_items_project_notion_id_fkey;
ALTER TABLE action_item_snapshots DROP CONSTRAINT IF EXISTS action_item_snapshots_project_notion_id_fkey;

ALTER TABLE projects DROP COLUMN IF EXISTS "notion_id";

CREATE TABLE IF NOT EXISTS project_notions
(
    id             uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at     TIMESTAMP(6),
    created_at     TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at     TIMESTAMP(6)     DEFAULT (NOW()),
    project_id       uuid NOT NULL,
    audit_notion_id uuid
);

ALTER TABLE project_notions ADD UNIQUE (audit_notion_id);
ALTER TABLE project_notions ADD CONSTRAINT project_notions_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);
ALTER TABLE audit_cycles ADD CONSTRAINT audit_cycles_project_notion_id_fkey FOREIGN KEY (project_id) REFERENCES project_notions (audit_notion_id);
ALTER TABLE audits ADD CONSTRAINT audits_project_notion_id_fkey FOREIGN KEY (project_id) REFERENCES project_notions (audit_notion_id);
ALTER TABLE action_items ADD CONSTRAINT action_items_project_notion_id_fkey FOREIGN KEY (project_id) REFERENCES project_notions (audit_notion_id);
ALTER TABLE action_item_snapshots ADD CONSTRAINT action_item_snapshots_project_notion_id_fkey FOREIGN KEY (project_id) REFERENCES project_notions (audit_notion_id);

-- +migrate Down
ALTER TABLE audit_cycles DROP CONSTRAINT IF EXISTS audit_cycles_project_notion_id_fkey;
ALTER TABLE audits DROP CONSTRAINT IF EXISTS audits_project_notion_id_fkey;
ALTER TABLE action_items DROP CONSTRAINT IF EXISTS action_items_project_notion_id_fkey;
ALTER TABLE action_item_snapshots DROP CONSTRAINT IF EXISTS action_item_snapshots_project_notion_id_fkey;
ALTER TABLE project_notions DROP CONSTRAINT IF EXISTS project_notions_project_id_fkey;
DROP TABLE IF EXISTS project_notions;
ALTER TABLE projects ADD COLUMN "notion_id" UUID;
ALTER TABLE projects ADD UNIQUE (notion_id);
ALTER TABLE audit_cycles ADD CONSTRAINT audit_cycles_project_notion_id_fkey FOREIGN KEY (project_id) REFERENCES projects (notion_id);
ALTER TABLE audits ADD CONSTRAINT audits_project_notion_id_fkey FOREIGN KEY (project_id) REFERENCES projects (notion_id);
ALTER TABLE action_items ADD CONSTRAINT action_items_project_notion_id_fkey FOREIGN KEY (project_id) REFERENCES projects (notion_id);
ALTER TABLE action_item_snapshots ADD CONSTRAINT action_item_snapshots_project_notion_id_fkey FOREIGN KEY (project_id) REFERENCES projects (notion_id);
