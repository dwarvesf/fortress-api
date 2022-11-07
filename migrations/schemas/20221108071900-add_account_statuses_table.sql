-- +migrate Up
CREATE TABLE "account_statuses" (
    "id" uuid PRIMARY KEY DEFAULT (uuid()),
    "deleted_at" TIMESTAMP(6),
    "created_at" TIMESTAMP(6) DEFAULT (now()),
    "updated_at" TIMESTAMP(6) DEFAULT (now()),
    "name" TEXT,
    "code" TEXT
);
ALTER TABLE employees
    ADD account_status uuid;
ALTER TABLE "employees" ADD CONSTRAINT employees_account_status_fkey FOREIGN KEY ("account_status") REFERENCES "account_statuses" ("id");
-- +migrate Down
ALTER TABLE employees
    DROP CONSTRAINT employees_account_status_fkey;
ALTER TABLE employees
    DROP column account_status;
DROP TABLE account_statuses;
