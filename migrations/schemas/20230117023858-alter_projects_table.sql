-- +migrate Up
ALTER TABLE "projects" ADD COLUMN "notion_id" UUID;
ALTER TABLE projects ADD UNIQUE (notion_id);

ALTER TABLE audit_cycles DROP CONSTRAINT IF EXISTS audit_cycles_project_id_fkey;
ALTER TABLE audits DROP CONSTRAINT IF EXISTS audits_project_id_fkey;
ALTER TABLE action_items DROP CONSTRAINT IF EXISTS action_items_project_id_fkey;
ALTER TABLE action_item_snapshots DROP CONSTRAINT IF EXISTS action_item_snapshots_project_id_fkey;

ALTER TABLE audit_cycles
    ADD CONSTRAINT audit_cycles_project_notion_id_fkey FOREIGN KEY (project_id) REFERENCES projects (notion_id);
ALTER TABLE audits
    ADD CONSTRAINT audits_project_notion_id_fkey FOREIGN KEY (project_id) REFERENCES projects (notion_id);
ALTER TABLE action_items
    ADD CONSTRAINT action_items_project_notion_id_fkey FOREIGN KEY (project_id) REFERENCES projects (notion_id);
ALTER TABLE action_item_snapshots
    ADD CONSTRAINT action_item_snapshots_project_notion_id_fkey FOREIGN KEY (project_id) REFERENCES projects (notion_id);


-- +migrate Down
ALTER TABLE audit_cycles
    ADD CONSTRAINT audit_cycles_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);
ALTER TABLE audits
    ADD CONSTRAINT audits_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);
ALTER TABLE action_items
    ADD CONSTRAINT action_items_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);
ALTER TABLE action_item_snapshots
    ADD CONSTRAINT action_item_snapshots_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);

ALTER TABLE audit_cycles DROP CONSTRAINT IF EXISTS audit_cycles_project_notion_id_fkey;
ALTER TABLE audits DROP CONSTRAINT IF EXISTS audits_project_notion_id_fkey;
ALTER TABLE action_items DROP CONSTRAINT IF EXISTS action_items_project_notion_id_fkey;
ALTER TABLE action_item_snapshots DROP CONSTRAINT IF EXISTS action_item_snapshots_project_notion_id_fkey;

ALTER TABLE "projects" DROP COLUMN "notion_id";
