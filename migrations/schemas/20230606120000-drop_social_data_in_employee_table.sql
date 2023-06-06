-- +migrate Up
DROP VIEW IF EXISTS vw_employees_recently_joined;

ALTER TABLE employees DROP COLUMN discord_id;
ALTER TABLE employees DROP COLUMN discord_name;
ALTER TABLE employees DROP COLUMN notion_id;
ALTER TABLE employees DROP COLUMN gitlab_id;
ALTER TABLE employees DROP COLUMN github_id;
ALTER TABLE employees DROP COLUMN notion_name;

CREATE OR REPLACE VIEW vw_employees_recently_joined AS
SELECT *
FROM employees
WHERE joined_date BETWEEN CURRENT_DATE - INTERVAL '7 days' AND CURRENT_DATE;

-- +migrate Down
ALTER TABLE employees ADD COLUMN discord_id TEXT;
ALTER TABLE employees ADD COLUMN discord_name TEXT;
ALTER TABLE employees ADD COLUMN notion_id TEXT;
ALTER TABLE employees ADD COLUMN gitlab_id TEXT;
ALTER TABLE employees ADD COLUMN github_id TEXT;
ALTER TABLE employees ADD COLUMN notion_name TEXT;
