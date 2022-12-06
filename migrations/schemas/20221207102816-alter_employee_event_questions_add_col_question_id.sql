
-- +migrate Up
ALTER TABLE employee_event_questions ADD IF NOT EXISTS question_id uuid;

ALTER TABLE employee_event_questions 
    ADD CONSTRAINT employee_event_questions_question_id_fkey FOREIGN KEY ("question_id") REFERENCES "questions" ("id");

-- +migrate Down
ALTER TABLE employee_event_questions 
    DROP CONSTRAINT IF EXISTS employee_event_questions_question_id_fkey;

ALTER TABLE employee_event_questions DROP COLUMN IF EXISTS question_id;
