
-- +migrate Up
ALTER TABLE "questions" RENAME "subtype" TO "subcategory";

ALTER TABLE "questions" ADD IF NOT EXISTS "category" TEXT;

ALTER TABLE "questions" ADD IF NOT EXISTS "order" INT;

ALTER TABLE "employee_event_questions" ADD IF NOT EXISTS "order" INT;
ALTER TABLE "employee_event_questions" ADD IF NOT EXISTS "type" TEXT;

ALTER TABLE employee_event_questions 
    DROP CONSTRAINT IF EXISTS employee_event_questions_question_id_fkey;

-- +migrate Down
ALTER TABLE employee_event_questions 
    ADD CONSTRAINT employee_event_questions_question_id_fkey FOREIGN KEY ("question_id") REFERENCES "questions" ("id");

ALTER TABLE "employee_event_questions" DROP COLUMN IF EXISTS "order";
ALTER TABLE "employee_event_questions" DROP COLUMN IF EXISTS "type";

ALTER TABLE "questions" DROP COLUMN IF EXISTS "order";

ALTER TABLE "questions" DROP COLUMN IF EXISTS "category";

ALTER TABLE "questions" RENAME "subcategory" TO "subtype";
