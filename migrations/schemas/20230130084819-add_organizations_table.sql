
-- +migrate Up
CREATE TABLE IF NOT EXISTS organizations (
	id uuid PRIMARY KEY DEFAULT(uuid()),
	deleted_at TIMESTAMP(6),
	created_at TIMESTAMP(6) DEFAULT(NOW()),
	updated_at TIMESTAMP(6) DEFAULT(NOW()),
	name TEXT,
	code TEXT,
	avatar TEXT
);

CREATE TABLE IF NOT EXISTS employee_organizations (
    id          uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at  TIMESTAMP(6),
    created_at  TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at  TIMESTAMP(6)     DEFAULT (NOW()),
    employee_id uuid NOT NULL,
    organization_id uuid NOT NULL
);

ALTER TABLE employee_organizations
    ADD CONSTRAINT employee_organizations_organizations_id_fkey FOREIGN KEY (organization_id) REFERENCES organizations (id);

ALTER TABLE employee_organizations
    ADD CONSTRAINT employee_organizations_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);

ALTER TABLE employees DROP COLUMN organization;
-- +migrate Down
ALTER TABLE employees
	ADD COLUMN organization TEXT;

DROP TABLE IF EXISTS employee_organizations;

DROP TABLE IF EXISTS organizations;

