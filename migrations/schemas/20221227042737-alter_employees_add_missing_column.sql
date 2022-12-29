
-- +migrate Up
ALTER TABLE "employees" ADD COLUMN "place_of_residence" TEXT;
ALTER TABLE "employees" ADD COLUMN "country" TEXT;
ALTER TABLE "employees" ADD COLUMN "city" TEXT;
ALTER TABLE "employees" ADD COLUMN "notion_email" TEXT;
ALTER TABLE "employees" ADD COLUMN "linkedin_name" TEXT;

ALTER TABLE "employees" ALTER COLUMN "gender" DROP NOT NULL;
ALTER TABLE "employees" ALTER COLUMN "avatar" DROP NOT NULL;

-- +migrate Down
ALTER TABLE "employees" DROP COLUMN "place_of_residence";
ALTER TABLE "employees" DROP COLUMN "country";
ALTER TABLE "employees" DROP COLUMN "city";
ALTER TABLE "employees" DROP COLUMN "notion_email";
ALTER TABLE "employees" DROP COLUMN "linkedin_name";

ALTER TABLE "employees" ALTER COLUMN "gender" SET NOT NULL;
ALTER TABLE "employees" ALTER COLUMN "avatar" SET NOT NULL;
