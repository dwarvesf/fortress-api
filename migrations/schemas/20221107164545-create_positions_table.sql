
-- +migrate Up
CREATE TABLE "positions" (
    "id" uuid PRIMARY KEY DEFAULT (uuid()),
    "deleted_at" TIMESTAMP(6),
    "created_at" TIMESTAMP(6) DEFAULT (now()),
    "updated_at" TIMESTAMP(6) DEFAULT (now()),
    "name" TEXT,
    "code" TEXT
);
ALTER TABLE employees
    ADD position uuid;
ALTER TABLE "employees" ADD CONSTRAINT fk_employee_positions FOREIGN KEY ("position") REFERENCES "positions" ("id");
-- +migrate Down
ALTER TABLE employees
    DROP CONSTRAINT fk_employee_positions;
ALTER TABLE employees
    DROP column position;
DROP TABLE positions;
