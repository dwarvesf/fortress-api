
-- +migrate Up
ALTER TABLE "seniorities" ADD COLUMN "level" INT;
ALTER TABLE "employees" ADD COLUMN "organization" TEXT;

-- +migrate Down
ALTER TABLE "seniorities" DROP COLUMN "level";
ALTER TABLE "employees" DROP COLUMN "organization";
