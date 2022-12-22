-- +migrate Up
ALTER TABLE "employees" ADD COLUMN "discord_name" TEXT;
ALTER TABLE "employees" ADD COLUMN "notion_name" TEXT;

-- +migrate Down
ALTER TABLE "employees" DROP COLUMN "discord_name";
ALTER TABLE "employees" DROP COLUMN "notion_name";
