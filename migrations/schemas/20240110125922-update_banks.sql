-- +migrate Up
ALTER TABLE "banks" ADD COLUMN "is_active" BOOLEAN NOT NULL DEFAULT FALSE;
-- +migrate Down
ALTER TABLE "banks" DROP COLUMN "is_active";
