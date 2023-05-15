
-- +migrate Up
DROP VIEW IF EXISTS "vw_idle_employees";
CREATE OR REPLACE VIEW "vw_idle_employees" AS
SELECT DISTINCT
	emp.full_name,
	o.name AS org_name
FROM
	employees emp
	JOIN employee_organizations eo ON emp.id = eo.employee_id
	JOIN organizations o ON o.id = eo.organization_id
WHERE
	emp.working_status NOT in('left', 'contractor')
	AND emp.id NOT in( SELECT DISTINCT
			mem.employee_id FROM projects p
		LEFT JOIN project_members mem ON mem.project_id = p.id
	WHERE
		p.status = 'active'
		AND mem.rate > 0
		AND mem.deployment_type <> 'shadow'
		AND mem.status <> 'inactive');

-- +migrate Down

DROP VIEW IF EXISTS "vw_idle_employees";
CREATE OR REPLACE VIEW "vw_idle_employees" AS
SELECT
  DISTINCT emp.*
FROM
  employees emp
  JOIN project_members pj ON pj.employee_id = emp.id
  AND emp.working_status <> 'left'
WHERE
  deployment_type <> 'shadow'
  OR rate = 0;
