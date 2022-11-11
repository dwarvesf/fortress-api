-- +migrate Up
CREATE TABLE "permissions" (
    "id" uuid PRIMARY KEY DEFAULT (uuid()),
    "deleted_at" TIMESTAMP(6),
    "created_at" TIMESTAMP(6) DEFAULT (now()),
    "updated_at" TIMESTAMP(6) DEFAULT (now()),
    "name" TEXT,
    "code" TEXT
);

CREATE TABLE "roles" (
    "id" uuid PRIMARY KEY DEFAULT (uuid()),
    "deleted_at" TIMESTAMP(6),
    "created_at" TIMESTAMP(6) DEFAULT (now()),
    "updated_at" TIMESTAMP(6) DEFAULT (now()),
    "name" TEXT,
    "code" TEXT
);

CREATE TABLE "role_permissions" (
    "id" uuid PRIMARY KEY DEFAULT (uuid()),
    "deleted_at" TIMESTAMP(6),
    "created_at" TIMESTAMP(6) DEFAULT (now()),
    "updated_at" TIMESTAMP(6) DEFAULT (now()),
    "role_id" uuid NOT NULL,
    "permission_id" uuid NOT NULL
);

CREATE TABLE "employee_roles" (
    "id" uuid PRIMARY KEY DEFAULT (uuid()),
    "deleted_at" TIMESTAMP(6),
    "created_at" TIMESTAMP(6) DEFAULT (now()),
    "updated_at" TIMESTAMP(6) DEFAULT (now()),
    "employee_id" uuid NOT NULL,
    "role_id" uuid NOT NULL
);

ALTER TABLE "role_permissions" ADD CONSTRAINT fk_role_permissions_roles FOREIGN KEY ("role_id") REFERENCES "roles" ("id");

ALTER TABLE "role_permissions" ADD CONSTRAINT fk_role_permissions_permissions FOREIGN KEY ("permission_id") REFERENCES "permissions" ("id");

ALTER TABLE "employee_roles" ADD CONSTRAINT fk_employee_roles_roles FOREIGN KEY ("role_id") REFERENCES "roles" ("id");

ALTER TABLE "employee_roles" ADD CONSTRAINT fk_employee_roles_employees FOREIGN KEY ("employee_id") REFERENCES "employees" ("id");

-- +migrate Down

ALTER TABLE "role_permissions" DROP CONSTRAINT fk_role_permissions_roles;

ALTER TABLE "role_permissions" DROP CONSTRAINT fk_role_permissions_permissions;

ALTER TABLE "employee_roles" DROP CONSTRAINT fk_employee_roles_roles;

ALTER TABLE "employee_roles" DROP CONSTRAINT fk_employee_roles_employees;

DROP TABLE employee_roles;

DROP TABLE role_permissions;

DROP TABLE roles;

DROP TABLE permissions;
