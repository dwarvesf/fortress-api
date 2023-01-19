
-- +migrate Up
-- +migrate StatementBegin
UPDATE questions
SET content = 'In the last two weeks, I have received recognition or praise for doing good work.'
WHERE id = '6c6bf3d6-46cd-46d3-9ea6-f382a5052588';

UPDATE employee_event_questions
SET content = 'In the last two weeks, I have received recognition or praise for doing good work.'
WHERE question_id = '6c6bf3d6-46cd-46d3-9ea6-f382a5052588';
-- +migrate StatementEnd
-- +migrate Down
SELECT TRUE;