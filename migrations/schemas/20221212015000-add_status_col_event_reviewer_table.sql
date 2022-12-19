-- +migrate Up
ALTER TABLE "employee_event_reviewers" DROP COLUMN "status";
ALTER TABLE "employee_event_reviewers" ADD COLUMN "author_status" TEXT;
ALTER TABLE "employee_event_reviewers" ADD COLUMN "reviewer_status" TEXT;
ALTER TABLE "employee_event_reviewers" ADD COLUMN "is_forced_done" BOOLEAN DEFAULT FALSE;

-- +migrate Down
ALTER TABLE "employee_event_reviewers" ADD COLUMN "status" TEXT;
ALTER TABLE "employee_event_reviewers" DROP COLUMN "author_status";
ALTER TABLE "employee_event_reviewers" DROP COLUMN "reviewer_status";
ALTER TABLE "employee_event_reviewers" DROP COLUMN "is_forced_done";
