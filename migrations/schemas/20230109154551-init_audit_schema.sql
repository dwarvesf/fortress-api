
-- +migrate Up

CREATE TYPE audit_status AS ENUM (
    'pending',
    'audited'
    );

CREATE TYPE audit_flag AS ENUM (
    'red',
    'yellow',
    'green',
    'none'
    );

CREATE TABLE IF NOT EXISTS audit_cycles
(
    id                  uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at          TIMESTAMP(6),
    created_at          TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at          TIMESTAMP(6)     DEFAULT (NOW()),
    project_id          uuid NOT NULL,
    notion_db_id        uuid NOT NULL,
    health_audit_id     uuid             DEFAULT NULL,
    process_audit_id    uuid             DEFAULT NULL,
    backend_audit_id    uuid             DEFAULT NULL,
    frontend_audit_id   uuid             DEFAULT NULL,
    system_audit_id     uuid             DEFAULT NULL,
    mobile_audit_id     uuid             DEFAULT NULL,
    blockchain_audit_id uuid             DEFAULT NULL,
    cycle               DECIMAL          DEFAULT 0,
    average_score       DECIMAL,
    status              audit_status,
    flag                audit_flag,
    quarter             TEXT             DEFAULT NULL,
    action_item_high    DECIMAL          DEFAULT 0,
    action_item_medium  DECIMAL          DEFAULT 0,
    action_item_low     DECIMAL          DEFAULT 0,
    sync_at             TIMESTAMP(6)     DEFAULT NULL
);

ALTER TABLE audit_cycles
    ADD CONSTRAINT audit_cycles_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);

CREATE TYPE audit_type AS ENUM (
    'engineering-health',
    'engineering-process',
    'frontend',
    'backend',
    'system',
    'mobile',
    'blockchain'
    );

CREATE TABLE IF NOT EXISTS audits
(
    id           uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at   TIMESTAMP(6),
    created_at   TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at   TIMESTAMP(6)     DEFAULT (NOW()),
    project_id   uuid NOT NULL,
    notion_db_id uuid NOT NULL,
    auditor_id   uuid NOT NULL,
    name         TEXT,
    type         audit_type,
    score        DECIMAL,
    status       audit_status,
    flag         audit_flag,
    action_item  DECIMAL,
    duration     DECIMAL,
    audited_at   TIMESTAMP(6)     DEFAULT NULL,
    sync_at      TIMESTAMP(6)     DEFAULT NULL
);

ALTER TABLE audits
    ADD CONSTRAINT audits_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);
ALTER TABLE audits
    ADD CONSTRAINT audits_auditor_id_fkey FOREIGN KEY (auditor_id) REFERENCES employees (id);

CREATE TYPE severity AS ENUM (
    'high',
    'medium',
    'low'
    );

CREATE TABLE IF NOT EXISTS audit_items
(
    id             uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at     TIMESTAMP(6),
    created_at     TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at     TIMESTAMP(6)     DEFAULT (NOW()),
    audit_id       uuid NOT NULL,
    notion_db_id uuid NOT NULL,
    name           TEXT,
    area           TEXT,
    requirements   TEXT,
    grade          DECIMAL,
    severity       severity,
    notes          TEXT,
    action_item_id uuid
);

ALTER TABLE audit_items
    ADD CONSTRAINT audit_items_audit_id_fkey FOREIGN KEY (audit_id) REFERENCES audits (id);


CREATE TYPE action_item_priority AS ENUM (
    'high',
    'medium',
    'low'
    );

CREATE TYPE action_item_status AS ENUM (
    'pending',
    'in-progress',
    'done'
    );

CREATE TABLE IF NOT EXISTS action_items
(
    id             uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at     TIMESTAMP(6),
    created_at     TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at     TIMESTAMP(6)     DEFAULT (NOW()),
    project_id     uuid,
    notion_db_id   uuid,
    pic_id uuid,
    audit_cycle_id uuid,
    name           TEXT,
    description    TEXT,
    need_help      BOOLEAN,
    priority       action_item_priority,
    status         action_item_status
);

ALTER TABLE action_items
    ADD CONSTRAINT action_items_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);
ALTER TABLE action_items
    ADD CONSTRAINT action_items_audit_cycle_id_fkey FOREIGN KEY (audit_cycle_id) REFERENCES audit_cycles (id);
ALTER TABLE action_items
    ADD CONSTRAINT action_items_pic_id_fkey FOREIGN KEY (pic_id) REFERENCES employees (id);

CREATE TABLE IF NOT EXISTS audit_action_items
(
    id             uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at     TIMESTAMP(6),
    created_at     TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at     TIMESTAMP(6)     DEFAULT (NOW()),
    audit_id       uuid NOT NULL,
    action_item_id uuid NOT NULL
);

ALTER TABLE audit_action_items
    ADD CONSTRAINT audit_action_items_audit_id_fkey FOREIGN KEY (audit_id) REFERENCES audits (id);
ALTER TABLE audit_action_items
    ADD CONSTRAINT audit_action_items_action_item_id_fkey FOREIGN KEY (action_item_id) REFERENCES action_items (id);

CREATE TABLE IF NOT EXISTS action_item_snapshots
(
    id             uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at     TIMESTAMP(6),
    created_at     TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at     TIMESTAMP(6)     DEFAULT (NOW()),
    project_id     uuid NOT NULL,
    audit_cycle_id uuid NOT NULL,
    high           DECIMAL          DEFAULT 0,
    medium         DECIMAL          DEFAULT 0,
    low            DECIMAL          DEFAULT 0
);

ALTER TABLE action_item_snapshots
    ADD CONSTRAINT action_item_snapshots_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);
ALTER TABLE action_item_snapshots
    ADD CONSTRAINT action_item_snapshots_audit_cycle_id_fkey FOREIGN KEY (audit_cycle_id) REFERENCES audit_cycles (id);

CREATE TABLE IF NOT EXISTS audit_participants
(
    id             uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at     TIMESTAMP(6),
    created_at     TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at     TIMESTAMP(6)     DEFAULT (NOW()),
    audit_id       uuid NOT NULL,
    employee_id    uuid NOT NULL
);

ALTER TABLE audit_participants
    ADD CONSTRAINT audit_participants_audit_id_fkey FOREIGN KEY (audit_id) REFERENCES audits (id);
ALTER TABLE audit_participants
    ADD CONSTRAINT audit_participants_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);
    
-- +migrate Down
ALTER TABLE audit_cycles DROP CONSTRAINT IF EXISTS audit_cycles_project_id_fkey;
ALTER TABLE audits DROP CONSTRAINT IF EXISTS audits_project_id_fkey;
ALTER TABLE audits DROP CONSTRAINT IF EXISTS audits_auditor_id_fkey;
ALTER TABLE audit_items DROP CONSTRAINT IF EXISTS audit_items_audit_id_fkey;
ALTER TABLE action_items DROP CONSTRAINT IF EXISTS action_items_project_id_fkey;
ALTER TABLE action_items DROP CONSTRAINT IF EXISTS action_items_audit_cycle_id_fkey;
ALTER TABLE action_items DROP CONSTRAINT IF EXISTS action_items_pic_id_fkey;
ALTER TABLE audit_action_items DROP CONSTRAINT IF EXISTS audit_action_items_audit_id_fkey;
ALTER TABLE audit_action_items DROP CONSTRAINT IF EXISTS audit_action_items_action_item_id_fkey;
ALTER TABLE action_item_snapshots DROP CONSTRAINT IF EXISTS action_item_snapshots_project_id_fkey;
ALTER TABLE action_item_snapshots DROP CONSTRAINT IF EXISTS action_item_snapshots_audit_cycle_id_fkey;
ALTER TABLE audit_participants DROP CONSTRAINT IF EXISTS audit_participants_audit_id_fkey;
ALTER TABLE audit_participants DROP CONSTRAINT IF EXISTS audit_participants_employee_id_fkey;

DROP TABLE IF EXISTS audit_cycles;
DROP TABLE IF EXISTS audits;
DROP TABLE IF EXISTS audit_items;
DROP TABLE IF EXISTS action_items;
DROP TABLE IF EXISTS audit_action_items;
DROP TABLE IF EXISTS action_item_snapshots;
DROP TABLE IF EXISTS audit_participants;

DROP TYPE IF EXISTS audit_status;
DROP TYPE IF EXISTS audit_flag;
DROP TYPE IF EXISTS audit_type;
DROP TYPE IF EXISTS severity;
DROP TYPE IF EXISTS action_item_status;
DROP TYPE IF EXISTS action_item_priority;
