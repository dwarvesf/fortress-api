-- +migrate Up
ALTER TABLE on_leave_requests ADD COLUMN notion_page_id VARCHAR(36);
CREATE INDEX idx_on_leave_requests_notion_page_id ON on_leave_requests(notion_page_id);

-- +migrate Down
DROP INDEX idx_on_leave_requests_notion_page_id;
ALTER TABLE on_leave_requests DROP COLUMN notion_page_id;
