-- +migrate Up
CREATE TABLE "employee_stacks" (
    "id"            uuid PRIMARY KEY DEFAULT (uuid()),
    "deleted_at"    TIMESTAMP(6),
    "created_at"    TIMESTAMP(6) DEFAULT (now()),
    "updated_at"    TIMESTAMP(6) DEFAULT (now()),
    "employee_id"   uuid NOT NULL,
    "stack_id"      uuid NOT NULL
);

ALTER TABLE "employee_stacks" ADD CONSTRAINT employee_stacks_stack_id_fkey FOREIGN KEY ("stack_id") REFERENCES "stacks" ("id");

ALTER TABLE "employee_stacks" ADD CONSTRAINT employee_stacks_employee_id_fkey FOREIGN KEY ("employee_id") REFERENCES "employees" ("id");

-- +migrate Down
ALTER TABLE "employee_stacks" DROP CONSTRAINT IF EXISTS employee_stacks_stack_id_fkey;

ALTER TABLE "employee_stacks" DROP CONSTRAINT IF EXISTS employee_stacks_employee_id_fkey;

DROP TABLE IF EXISTS employee_stacks;
