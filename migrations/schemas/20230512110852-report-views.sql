-- +migrate Up

CREATE VIEW "vw_idle_employees" AS
SELECT
  DISTINCT emp.*
FROM
  employees emp
  JOIN project_members pj ON pj.employee_id = emp.id
  AND emp.working_status <> 'left'
WHERE
  deployment_type <> 'shadow'
  OR rate = 0;

CREATE VIEW "vw_subscribers_last_7days" AS
SELECT 
  *
FROM
  audiences
WHERE now() :: date - created_at :: date <= 7;

CREATE OR REPLACE VIEW "vw_icy_earning_by_team_monthly" AS 
WITH weekly_earning AS (
  SELECT
    date_trunc('month', txn_time) AS "month",
    date_trunc('week', txn_time) AS "week",
    category,
    SUM(amount) AS "amount"
  FROM
    icy_transactions
  GROUP BY
    date_trunc('month', txn_time),
    date_trunc('week', txn_time),
    category
  ORDER BY
    date_trunc('week', txn_time) DESC,
    SUM(amount) DESC
)
SELECT
  to_char(date_trunc('month', txn_time), 'yyyy-mm') as "period",
  t.category as "team",
  SUM(t.amount) as "amount",
  AVG(w.amount) AS "avg_earning_weekly"
FROM
  icy_transactions t
  LEFT JOIN weekly_earning w ON w.category = t.category
  AND w.month = date_trunc('month', txn_time)
GROUP BY
  date_trunc('month', txn_time),
  t.category
ORDER BY
  date_trunc('month', txn_time) DESC,
  SUM(t.amount) DESC;

CREATE OR REPLACE VIEW "vw_icy_earning_by_team_all_time" AS WITH monthly_earning AS (
  SELECT
    date_trunc('month', txn_time) AS "month",
    category,
    SUM(amount) AS "amount"
  FROM
    icy_transactions
  GROUP BY
    date_trunc('month', txn_time),
    category
  ORDER BY
    date_trunc('month', txn_time) DESC,
    SUM(amount) DESC
),
weekly_earning AS (
  SELECT
    date_trunc('month', txn_time) AS "month",
    date_trunc('week', txn_time) AS "week",
    category,
    SUM(amount) AS "amount"
  FROM
    icy_transactions
  GROUP BY
    date_trunc('month', txn_time),
    date_trunc('week', txn_time),
    category
  ORDER BY
    date_trunc('week', txn_time) DESC,
    SUM(amount) DESC
)
SELECT
  m.category AS "team",
  SUM(t.amount) AS "amount",
  AVG(m.amount) AS "avg_earning_monthy",
  AVG(w.amount) AS "avg_earning_weekly"
FROM
  icy_transactions t
  LEFT JOIN monthly_earning m ON m.category = t.category
  LEFT JOIN weekly_earning w ON w.category = t.category
  AND w.month = m.month
GROUP BY
  m.category
ORDER BY
  SUM(t.amount) DESC;

CREATE OR REPLACE VIEW "vw_icy_earning_by_team_weekly" AS
SELECT
  to_char(date_trunc('week', txn_time), 'yyyy-mm-dd') AS "period",
  category AS "team",
  SUM(t.amount) AS "amount"
FROM
  icy_transactions t
GROUP BY
  date_trunc('week', txn_time),
  category
ORDER BY
  date_trunc('week', txn_time) DESC,
  SUM(t.amount) DESC;

CREATE OR REPLACE VIEW "vw_icy_employee_dashboard" AS
SELECT
  t.dest_employee_id as employee_id,
  e.full_name,
  e.team_email,
  e.personal_email,
  SUM(t.amount) AS "total_earned"
FROM
  icy_transactions t
  LEFT JOIN employees e ON e.id = t.dest_employee_id
GROUP BY
  t.dest_employee_id,
  e.full_name,
  e.team_email,
  e.personal_email
ORDER BY
  SUM(t.amount) DESC;

-- +migrate Down

DROP VIEW IF EXISTS "vw_subscribers_last_7days";

DROP VIEW IF EXISTS "vw_icy_earning_by_team_weekly";

DROP VIEW IF EXISTS "vw_icy_earning_by_team_monthly";

DROP VIEW IF EXISTS "vw_icy_earning_by_team_all_time";

DROP VIEW IF EXISTS "vw_icy_employee_dashboard";

DROP VIEW IF EXISTS "vw_idle_employees";
