
-- +migrate Up
CREATE TABLE "tech_stacks" (
    "id" uuid PRIMARY KEY DEFAULT (uuid()),
    "deleted_at" TIMESTAMP(6),
    "created_at" TIMESTAMP(6) DEFAULT (now()),
    "updated_at" TIMESTAMP(6) DEFAULT (now()),
    "name" TEXT,
    "code" TEXT
);

CREATE TABLE "employee_tech_stacks" (
    "id" uuid PRIMARY KEY DEFAULT (uuid()),
    "deleted_at" TIMESTAMP(6),
    "created_at" TIMESTAMP(6) DEFAULT (now()),
    "updated_at" TIMESTAMP(6) DEFAULT (now()),
    "employee_id" uuid NOT NULL,
    "tech_stack_id" uuid NOT NULL
);

ALTER TABLE "employee_tech_stacks" ADD CONSTRAINT employee_tech_stacks_tech_stack_id_fkey FOREIGN KEY ("tech_stack_id") REFERENCES "tech_stacks" ("id");

ALTER TABLE "employee_tech_stacks" ADD CONSTRAINT employee_tech_stacks_employee_id_fkey FOREIGN KEY ("employee_id") REFERENCES "employees" ("id");

-- +migrate Down
ALTER TABLE "employee_tech_stacks" DROP CONSTRAINT IF EXISTS employee_tech_stacks_tech_stack_id_fkey;

ALTER TABLE "employee_tech_stacks" DROP CONSTRAINT IF EXISTS employee_tech_stacks_employee_id_fkey;

DROP TABLE IF EXISTS employee_tech_stacks;

DROP TABLE IF EXISTS tech_stacks;
