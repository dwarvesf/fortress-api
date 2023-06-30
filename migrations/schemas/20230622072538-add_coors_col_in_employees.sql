-- +migrate Up
DROP VIEW IF EXISTS vw_employees_recently_joined;
ALTER TABLE employees ADD COLUMN lat TEXT DEFAULT NULL;
ALTER TABLE employees ADD COLUMN long TEXT DEFAULT NULL;

CREATE OR REPLACE VIEW vw_employees_recently_joined AS
SELECT *
FROM employees
WHERE joined_date BETWEEN CURRENT_DATE - INTERVAL '7 days' AND CURRENT_DATE;
-- +migrate Down
ALTER TABLE employees DROP COLUMN lat CASCADE;
ALTER TABLE employees DROP COLUMN long CASCADE;
