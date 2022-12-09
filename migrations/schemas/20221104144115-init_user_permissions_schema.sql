-- +migrate Up
CREATE TABLE IF NOT EXISTS permissions (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (now()),
    updated_at TIMESTAMP(6)     DEFAULT (now()),
    name       TEXT,
    code       TEXT
);

CREATE TABLE IF NOT EXISTS roles (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (now()),
    updated_at TIMESTAMP(6)     DEFAULT (now()),
    name       TEXT,
    code       TEXT
);

CREATE TABLE IF NOT EXISTS role_permissions (
    id            uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at    TIMESTAMP(6),
    created_at    TIMESTAMP(6)     DEFAULT (now()),
    updated_at    TIMESTAMP(6)     DEFAULT (now()),
    role_id       uuid NOT NULL,
    permission_id uuid NOT NULL
);

CREATE TABLE IF NOT EXISTS employee_roles (
    id          uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at  TIMESTAMP(6),
    created_at  TIMESTAMP(6)     DEFAULT (now()),
    updated_at  TIMESTAMP(6)     DEFAULT (now()),
    employee_id uuid NOT NULL,
    role_id     uuid NOT NULL
);

ALTER TABLE role_permissions
    ADD CONSTRAINT role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles (id);

ALTER TABLE role_permissions
    ADD CONSTRAINT role_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES permissions (id);

ALTER TABLE employee_roles
    ADD CONSTRAINT employee_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles (id);

ALTER TABLE employee_roles
    ADD CONSTRAINT employee_roles_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);

-- +migrate Down

ALTER TABLE role_permissions DROP CONSTRAINT IF EXISTS role_permissions_role_id_fkey;

ALTER TABLE role_permissions DROP CONSTRAINT IF EXISTS role_permissions_permission_id_fkey;

ALTER TABLE employee_roles DROP CONSTRAINT IF EXISTS employee_roles_role_id_fkey;

ALTER TABLE employee_roles DROP CONSTRAINT  IF EXISTS employee_roles_employee_id_fkey;

DROP TABLE IF EXISTS employee_roles;

DROP TABLE IF EXISTS role_permissions;

DROP TABLE IF EXISTS roles;

DROP TABLE IF EXISTS permissions;
