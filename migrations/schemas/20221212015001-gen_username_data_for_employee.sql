-- +migrate Up
ALTER TABLE "employees" DROP CONSTRAINT IF EXISTS employees_username_key;
-- +migrate StatementBegin
UPDATE employees
SET username = SPLIT_PART(team_email, '@', 1);
-- +migrate StatementEnd
-- +migrate Down
ALTER TABLE "employees" DROP CONSTRAINT IF EXISTS employees_username_key;
UPDATE employees
SET username = NULL;

ALTER TABLE employees ADD UNIQUE (username);
