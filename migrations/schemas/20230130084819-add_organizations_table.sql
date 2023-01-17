
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

INSERT INTO "public"."organizations" ("id", "deleted_at", "created_at", "updated_at", "name", "code", "avatar") VALUES
('31fdf38f-77c0-4c06-b530-e2be8bc297e0', NULL, '2023-01-19 11:13:13.487168', '2023-01-19 11:13:13.487168', 'Dwarves Foundation', 'dwarves-foundation', NULL),
('e4725383-943a-468a-b0cd-ce249c573cf7', NULL, '2023-01-19 11:13:13.487168', '2023-01-19 11:13:13.487168', 'Console Labs', 'console-labs', NULL);


INSERT INTO employee_organizations (employee_id, organization_id)
SELECT
	e.id,
	o.id
FROM
	employees e
	JOIN organizations o ON e.organization = o.code;

ALTER TABLE employees DROP COLUMN organization;

-- +migrate Down

ALTER TABLE employees
	ADD COLUMN organization TEXT;

UPDATE
	employees e
SET
	organization = tmp.code
FROM (
	SELECT
		o.code,
		e.id
	FROM
		employees e
		JOIN employee_organizations eo ON e.id = eo.employee_id
		JOIN organizations o ON eo.organization_id = o.id) tmp
WHERE
	e.id = tmp.id;

DROP TABLE IF EXISTS employee_organizations;

DROP TABLE IF EXISTS organizations;

