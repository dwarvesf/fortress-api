-- +migrate Up
ALTER TABLE "roles" ADD COLUMN "level" INT;

-- +migrate Down
ALTER TABLE "roles" DROP COLUMN "level";