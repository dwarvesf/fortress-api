
-- +migrate Up
ALTER TABLE "employees" ADD COLUMN "shelter_address" TEXT;
ALTER TABLE "employees" ADD COLUMN "permanent_address" TEXT;
ALTER TABLE "employees" ADD COLUMN "country" TEXT;
ALTER TABLE "employees" ADD COLUMN "city" TEXT;
ALTER TABLE "employees" ADD COLUMN "notion_email" TEXT;
ALTER TABLE "employees" ADD COLUMN "linkedin_name" TEXT;

-- +migrate Down
ALTER TABLE "employees" DROP COLUMN "shelter_address";
ALTER TABLE "employees" DROP COLUMN "permanent_address";
ALTER TABLE "employees" DROP COLUMN "country";
ALTER TABLE "employees" DROP COLUMN "city";
ALTER TABLE "employees" DROP COLUMN "notion_email";
ALTER TABLE "employees" DROP COLUMN "linkedin_name";
