
-- +migrate Up
ALTER TABLE employees
    ADD line_manager_id uuid;
ALTER TABLE "employees" ADD CONSTRAINT employee_line_manager_id_fkey FOREIGN KEY ("line_manager_id") REFERENCES "employees" ("id");

-- +migrate Down
ALTER TABLE "employees" DROP CONSTRAINT IF EXISTS employee_line_manager_id_fkey;
ALTER TABLE employees
    DROP column IF EXISTS line_manager_id;
