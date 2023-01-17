
-- +migrate Up
-- +migrate StatementBegin
UPDATE questions
SET content = 'Does this employee effectively communicate with others?'
WHERE id = 'da5dbdd5-8e1e-4ae7-8bb8-ab007f2580aa';
UPDATE questions
SET content = 'How effective of a leader is this person, either through direct management or influence?'
WHERE id = '7d95e035-81d6-49d7-bed4-3a83bf2e34d6';
UPDATE questions
SET content = 'Does this person find creative solutions, and own the solution to problems? Are they proactive or reactive?'
WHERE id = 'd36e84c5-d7a4-4d5f-ada1-f6b9ddb58f51';
UPDATE questions
SET content = 'How would you rate the quality of the employee''s work?'
WHERE id = 'f03432ba-c024-467e-8059-a5bb2b7f783d';
UPDATE questions
SET content = 'How well does this person set and meet deadlines?'
WHERE id = 'd2bb48c1-e8d6-4946-a372-8499907b7328';
UPDATE questions
SET content = 'How well does this person embody our culture?'
WHERE id = 'be86ce52-803b-403f-b059-1a69492fe3d4';
UPDATE questions
SET content = 'If you could give this person one piece of constructive advice to make them more effective in their role, what would you say?'
WHERE id = '51eab8c7-61ba-4c56-be39-b72eb6b89a52';

UPDATE employee_event_questions
SET content = 'Does this employee effectively communicate with others?'
WHERE question_id = 'da5dbdd5-8e1e-4ae7-8bb8-ab007f2580aa';
UPDATE employee_event_questions
SET content = 'How effective of a leader is this person, either through direct management or influence?'
WHERE question_id = '7d95e035-81d6-49d7-bed4-3a83bf2e34d6';
UPDATE employee_event_questions
SET content = 'Does this person find creative solutions, and own the solution to problems? Are they proactive or reactive?'
WHERE question_id = 'd36e84c5-d7a4-4d5f-ada1-f6b9ddb58f51';
UPDATE employee_event_questions
SET content = 'How would you rate the quality of the employee''s work?'
WHERE question_id = 'f03432ba-c024-467e-8059-a5bb2b7f783d';
UPDATE employee_event_questions
SET content = 'How well does this person set and meet deadlines?'
WHERE question_id = 'd2bb48c1-e8d6-4946-a372-8499907b7328';
UPDATE employee_event_questions
SET content = 'How well does this person embody our culture?'
WHERE question_id = 'be86ce52-803b-403f-b059-1a69492fe3d4';
UPDATE employee_event_questions
SET content = 'If you could give this person one piece of constructive advice to make them more effective in their role, what would you say?'
WHERE question_id = '51eab8c7-61ba-4c56-be39-b72eb6b89a52'
-- +migrate StatementEnd
-- +migrate Down
SELECT TRUE;