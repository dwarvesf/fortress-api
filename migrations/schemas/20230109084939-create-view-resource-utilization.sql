
-- +migrate Up
CREATE VIEW resource_utilization AS
SELECT d.d AS "date", 
		e.id AS employee_id, 
		p.id AS project_id,
		p."type" AS "project_type",
		pm.deployment_type,
		pm.status,
		pm.joined_date,
		pm.left_date 
FROM employees e 
	LEFT JOIN project_members pm ON pm.employee_id = e.id
	LEFT JOIN projects p ON pm.project_id = p.id, 
	generate_series(
		date_trunc('month', CURRENT_DATE) - INTERVAL '11 month', 
		date_trunc('month', CURRENT_DATE), 
		'1 month'
	) d
WHERE e."working_status" = 'full-time'
ORDER BY d.d;

-- +migrate Down
DROP VIEW IF EXISTS resource_utilization;
