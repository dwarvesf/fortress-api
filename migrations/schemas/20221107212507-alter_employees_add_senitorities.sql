-- +migrate Up
CREATE TABLE IF NOT EXISTS "seniorities" (
    "id" uuid PRIMARY KEY DEFAULT (uuid()),
    "deleted_at" TIMESTAMP(6),
    "created_at" TIMESTAMP(6) DEFAULT (now()),
    "updated_at" TIMESTAMP(6) DEFAULT (now()),
    "name" TEXT,
    "code" TEXT
);
ALTER TABLE employees ADD seniority_id uuid;
ALTER TABLE "employees" ADD CONSTRAINT employees_senioriry_id_fkey FOREIGN KEY ("seniority_id") REFERENCES "seniorities" ("id");

-- +migrate Down
ALTER TABLE employees DROP CONSTRAINT employees_senioriry_id_fkey;
ALTER TABLE employees DROP column seniority_id;
DROP TABLE IF EXISTS seniorities;
