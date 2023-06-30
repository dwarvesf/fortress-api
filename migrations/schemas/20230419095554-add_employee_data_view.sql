-- +migrate Up
CREATE OR REPLACE VIEW vw_employee_count_by_chapter AS
    WITH chapter_data as (
        SELECT c.id,
        CASE
            WHEN c.code IN ('backend', 'web', 'mobile', 'blockchain', 'devops', 'data') THEN 'Developer'
            WHEN c.code = 'qa' THEN 'QA'
            WHEN c.code = 'design' THEN 'Designer'
            WHEN c.code IN ('operations', 'pm') THEN 'Operation'
            WHEN c.code = 'sales' THEN 'Sales'
            ELSE c.code
        END AS chapter_group_name,
        c.name,
        c.code,
        c.lead_id
        FROM chapters c
    )
    SELECT
        chapter_data.chapter_group_name,
        COUNT(DISTINCT (ec.employee_id)) AS employee_count
    FROM chapter_data
         LEFT JOIN employee_chapters ec ON ec.chapter_id = chapter_data.id
         LEFT JOIN employees e ON ec.employee_id = e.id
    WHERE e.working_status IN ('probation', 'full-time', 'contractor')
    GROUP BY chapter_data.chapter_group_name
    ORDER BY employee_count DESC
;

CREATE OR REPLACE VIEW vw_employees_recently_joined AS
    SELECT *
    FROM employees
    WHERE joined_date BETWEEN CURRENT_DATE - INTERVAL '7 days' AND CURRENT_DATE
;

CREATE OR REPLACE VIEW vw_employees_project_charge_rates AS
    WITH time as (SELECT NOW() as now)
    SELECT e.id as employee_id, e.full_name, pm.project_id, pm.rate, ba.currency_id
    FROM time,
         project_members pm
             JOIN projects p ON pm.project_id = p.id
             JOIN employees e ON pm.employee_id = e.id
             JOIN bank_accounts ba on p.bank_account_id = ba.id
    WHERE e.working_status IN ('probation', 'full-time', 'contractor')
      AND p.type = 'time-material'
      AND p.status = 'active'
      AND (pm.start_date <= time.now::DATE AND (pm.end_date IS NULL OR pm.end_date <= time.now::DATE))
      AND pm.rate > 0
      AND pm.deployment_type = 'official'
;

-- +migrate Down
DROP VIEW vw_employee_count_by_chapter;
DROP VIEW vw_employees_project_charge_rates;
