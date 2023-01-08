
-- +migrate Up
ALTER TABLE "seniorities" ADD COLUMN "level" INT;

-- +migrate Down
ALTER TABLE "seniorities" DROP COLUMN "level";
