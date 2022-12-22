-- +migrate Up
ALTER TABLE "employee_event_reviewers" DROP COLUMN "status";
ALTER TABLE "employee_event_reviewers" ADD COLUMN "author_status" TEXT;
ALTER TABLE "employee_event_reviewers" ADD COLUMN "reviewer_status" TEXT;
ALTER TABLE "employee_event_reviewers" ADD COLUMN "is_forced_done" BOOLEAN DEFAULT FALSE;
ALTER TABLE "employee_event_questions" ADD COLUMN "domain" TEXT;
ALTER TABLE "questions" ADD COLUMN "domain" TEXT;
ALTER TABLE "projects" ADD COLUMN "allows_sending_survey" BOOLEAN DEFAULT FALSE;
ALTER TABLE "employees" ADD COLUMN "discord_name" TEXT;
ALTER TABLE "employees" ADD COLUMN "notion_name" TEXT;

-- +migrate Down
ALTER TABLE "employee_event_reviewers" ADD COLUMN "status" TEXT;
ALTER TABLE "employee_event_reviewers" DROP COLUMN "author_status";
ALTER TABLE "employee_event_reviewers" DROP COLUMN "reviewer_status";
ALTER TABLE "employee_event_reviewers" DROP COLUMN "is_forced_done";
ALTER TABLE "employee_event_questions" DROP COLUMN "domain";
ALTER TABLE "questions" DROP COLUMN "domain";
ALTER TABLE "projects" DROP COLUMN "allows_sending_survey";
ALTER TABLE "employees" DROP COLUMN "discord_name";
ALTER TABLE "employees" DROP COLUMN "notion_name";
