-- +migrate Up
CREATE TYPE work_unit_types AS ENUM (
  'development',
  'management',
  'training',
  'learning'
);

CREATE TYPE work_unit_statuses AS ENUM (
  'active',
  'archived'
);

CREATE TABLE IF NOT EXISTS work_units (
    id               uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at       TIMESTAMP(6),
    created_at       TIMESTAMP(6)     DEFAULT (now()),
    updated_at       TIMESTAMP(6)     DEFAULT (now()),
    name             TEXT,
    status           work_unit_statuses,
    type             work_unit_types,
    source_url       TEXT,
    project_id       uuid NOT NULL,
    source_metadata  JSONB            DEFAULT '[]'::JSONB
);

CREATE TABLE IF NOT EXISTS work_unit_stacks (
    id                uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at        TIMESTAMP(6),
    created_at        TIMESTAMP(6)     DEFAULT (now()),
    updated_at        TIMESTAMP(6)     DEFAULT (now()),
    stack_id          uuid NOT NULL,
    work_unit_id      uuid NOT NULL
);

CREATE TABLE IF NOT EXISTS work_unit_members (
    id                uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at        TIMESTAMP(6),
    created_at        TIMESTAMP(6)     DEFAULT (now()),
    updated_at        TIMESTAMP(6)     DEFAULT (now()),
    joined_date       DATE NOT NULL,
    left_date         DATE             DEFAULT NULL,
    status            TEXT,
    project_id        uuid NOT NULL,
    employee_id       uuid NOT NULL,
    work_unit_id      uuid NOT NULL
);

ALTER TABLE work_units
    ADD CONSTRAINT work_units_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);

ALTER TABLE work_unit_stacks
    ADD CONSTRAINT work_unit_stacks_stack_id_fkey FOREIGN KEY (stack_id) REFERENCES stacks (id);

ALTER TABLE work_unit_stacks
    ADD CONSTRAINT work_unit_stacks_work_unit_id_fkey FOREIGN KEY (work_unit_id) REFERENCES work_units (id);

ALTER TABLE work_unit_members
    ADD CONSTRAINT work_unit_members_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);

ALTER TABLE work_unit_members
    ADD CONSTRAINT work_unit_members_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);

ALTER TABLE work_unit_members
    ADD CONSTRAINT work_unit_members_work_unit_id_fkey FOREIGN KEY (work_unit_id) REFERENCES work_units (id);

-- +migrate Down

ALTER TABLE work_units DROP CONSTRAINT IF EXISTS work_units_project_id_fkey;
ALTER TABLE work_unit_stacks DROP CONSTRAINT IF EXISTS work_unit_stacks_stack_id_fkey;
ALTER TABLE work_unit_stacks DROP CONSTRAINT IF EXISTS work_unit_stacks_work_unit_id_fkey;
ALTER TABLE work_unit_members DROP CONSTRAINT  IF EXISTS work_unit_members_project_id_fkey;
ALTER TABLE work_unit_members DROP CONSTRAINT IF EXISTS work_unit_members_employee_id_fkey;
ALTER TABLE work_unit_members DROP CONSTRAINT  IF EXISTS work_unit_members_work_unit_id_fkey;

DROP TABLE IF EXISTS work_unit_members;
DROP TABLE IF EXISTS work_unit_stacks;
DROP TABLE IF EXISTS work_units;

DROP TYPE IF EXISTS work_unit_statuses;
DROP TYPE IF EXISTS work_unit_types;