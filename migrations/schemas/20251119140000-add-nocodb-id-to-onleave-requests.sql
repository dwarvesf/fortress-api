-- +migrate Up
ALTER TABLE on_leave_requests ADD COLUMN nocodb_id INTEGER;
CREATE INDEX idx_on_leave_requests_nocodb_id ON on_leave_requests(nocodb_id);

-- +migrate Down
DROP INDEX idx_on_leave_requests_nocodb_id;
ALTER TABLE on_leave_requests DROP COLUMN nocodb_id;
