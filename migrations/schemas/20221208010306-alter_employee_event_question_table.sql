
-- +migrate Up
ALTER TABLE "employee_event_questions" RENAME "answers" TO "answer";

ALTER TABLE "employee_event_questions" ADD IF NOT EXISTS "event_id" UUID;
ALTER TABLE "employee_event_questions" 
    ADD CONSTRAINT "employee_event_questions_event_id_fkey" FOREIGN KEY ("event_id") REFERENCES "feedback_events" ("id");

ALTER TABLE "employee_event_reviewers" ADD IF NOT EXISTS "event_id" UUID;
ALTER TABLE "employee_event_reviewers" 
    ADD CONSTRAINT "employee_event_reviewers_event_id_fkey" FOREIGN KEY ("event_id") REFERENCES "feedback_events" ("id");

-- +migrate Down
ALTER TABLE "employee_event_questions" 
    DROP CONSTRAINT IF EXISTS "employee_event_questions_event_id_fkey";
ALTER TABLE "employee_event_questions" DROP COLUMN IF EXISTS "event_id";

ALTER TABLE "employee_event_reviewers" 
    DROP CONSTRAINT IF EXISTS "employee_event_reviewers_event_id_fkey";
ALTER TABLE "employee_event_reviewers" DROP COLUMN IF EXISTS "event_id";

ALTER TABLE "employee_event_questions" RENAME "answer" TO "answers";

