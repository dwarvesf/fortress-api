-- +migrate Up
CREATE TYPE project_types AS ENUM (
  'time-material',
  'fixed-cost',
  'dwarves'
);

CREATE TYPE  project_head_positions AS ENUM (
  'delivery-manager',
  'account-manager',
  'sale-person',
  'technical-lead'
);

CREATE TYPE  project_statuses AS ENUM (
  'on-boarding',
  'active',
  'closed',
  'paused'
);

CREATE TYPE  deployment_types AS ENUM (
  'shadow',
  'official'
);

CREATE TYPE project_member_statuses AS ENUM (
  'pending',
  'on-boarding',
  'active',
  'inactive'
);

CREATE TABLE IF NOT EXISTS projects (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at TIMESTAMP(6)     DEFAULT (NOW()),
    name       text,
    type       project_types,
    start_date DATE,
    end_date   DATE,
    status     project_statuses
);

CREATE TABLE IF NOT EXISTS project_slots (
    id               uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at       TIMESTAMP(6),
    created_at       TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at       TIMESTAMP(6)     DEFAULT (NOW()),
    project_id       uuid NOT NULL,
    seniority_id     uuid NOT NULL,
    upsell_person_id uuid             DEFAULT NULL,
    position         TEXT,
    deployment_type  deployment_types,
    rate             DECIMAL,
    discount         DECIMAL,
    status           TEXT
);

CREATE TABLE IF NOT EXISTS project_members (
    id               uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at       TIMESTAMP(6),
    created_at       TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at       TIMESTAMP(6)     DEFAULT (NOW()),
    project_id       uuid NOT NULL,
    project_slot_id  uuid NOT NULL,
    employee_id      uuid NOT NULL,
    seniority_id     uuid NOT NULL,
    joined_date      DATE NOT NULL,
    left_date        DATE             DEFAULT NULL,
    position         TEXT,
    rate             DECIMAL,
    discount         DECIMAL,
    status           project_member_statuses,
    deployment_type  deployment_types,
    upsell_person_id uuid             DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS project_heads (
    id              uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at      TIMESTAMP(6),
    created_at      TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at      TIMESTAMP(6)     DEFAULT (NOW()),
    project_id      uuid NOT NULL,
    employee_id     uuid NOT NULL,
    joined_date     DATE NOT NULL,
    left_date       DATE             DEFAULT NULL,
    commission_rate DECIMAL,
    position        project_head_positions
);

CREATE TABLE IF NOT EXISTS project_stacks (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at TIMESTAMP(6)     DEFAULT (NOW()),
    project_id uuid NOT NULL,
    stack_id   uuid NOT NULL
);

CREATE TABLE IF NOT EXISTS stacks (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at TIMESTAMP(6)     DEFAULT (NOW()),
    name       TEXT,
    code       TEXT,
    avatar     TEXT
);

ALTER TABLE project_slots
    ADD CONSTRAINT project_slots_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);

ALTER TABLE project_slots
    ADD CONSTRAINT project_slots_seniority_id_fkey FOREIGN KEY (seniority_id) REFERENCES seniorities (id);

ALTER TABLE project_slots
    ADD CONSTRAINT project_slots_upsell_person_id_fkey FOREIGN KEY (upsell_person_id) REFERENCES employees (id);

ALTER TABLE project_members
    ADD CONSTRAINT project_members_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);

ALTER TABLE project_members
    ADD CONSTRAINT project_members_project_slot_id_fkey FOREIGN KEY (project_slot_id) REFERENCES project_slots (id);

ALTER TABLE project_members
    ADD CONSTRAINT project_members_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);

ALTER TABLE project_members
    ADD CONSTRAINT project_members_upsell_person_id_fkey FOREIGN KEY (upsell_person_id) REFERENCES employees (id);

ALTER TABLE project_members
    ADD CONSTRAINT project_members_seniority_id_fkey FOREIGN KEY (seniority_id) REFERENCES seniorities (id);

ALTER TABLE project_heads
    ADD CONSTRAINT project_heads_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);

ALTER TABLE project_heads
    ADD CONSTRAINT project_heads_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);

ALTER TABLE project_stacks
    ADD CONSTRAINT project_stacks_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);

ALTER TABLE project_stacks
    ADD CONSTRAINT project_stacks_stack_id_fkey FOREIGN KEY (stack_id) REFERENCES stacks (id);

-- +migrate Down
ALTER TABLE project_members DROP CONSTRAINT project_members_project_id_fkey;
ALTER TABLE project_members DROP CONSTRAINT project_members_employee_id_fkey;
ALTER TABLE project_members DROP CONSTRAINT project_members_upsell_person_id_fkey;
ALTER TABLE project_members DROP CONSTRAINT project_members_project_slot_id_fkey;
ALTER TABLE project_members DROP CONSTRAINT project_members_seniority_id_fkey;
ALTER TABLE project_slots DROP CONSTRAINT project_slots_project_id_fkey;
ALTER TABLE project_slots DROP CONSTRAINT project_slots_seniority_id_fkey;
-- ALTER TABLE project_slots DROP CONSTRAINT project_slots_upsell_person_id_fkey;
ALTER TABLE project_heads DROP CONSTRAINT project_heads_project_id_fkey;
ALTER TABLE project_heads DROP CONSTRAINT project_heads_employee_id_fkey;
ALTER TABLE project_stacks DROP CONSTRAINT project_stacks_project_id_fkey;
ALTER TABLE project_stacks DROP CONSTRAINT project_stacks_stack_id_fkey;

DROP TABLE IF EXISTS project_stacks;
DROP TABLE IF EXISTS stacks;
DROP TABLE IF EXISTS project_slots;
DROP TABLE IF EXISTS project_heads;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;
DROP TYPE IF EXISTS project_types;
DROP TYPE IF EXISTS project_head_positions;
DROP TYPE IF EXISTS project_statuses;
DROP TYPE IF EXISTS deployment_types;
DROP TYPE IF EXISTS project_member_statuses;
