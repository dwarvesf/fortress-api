-- +migrate Up
ALTER TABLE employees
    DROP CONSTRAINT employees_account_status_fkey;
ALTER TABLE employees
    DROP column account_status;

DROP TABLE account_statuses;

CREATE TYPE account_status as ENUM('onboarding', 'probation', 'active', 'on-leave');

ALTER TABLE employees
    ADD account_status account_status;

-- +migrate Down
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

drop type account_status;

ALTER TABLE employees
    DROP column IF EXISTS account_status;
