-- +migrate Up
ALTER TABLE icy_transactions ALTER COLUMN dest_employee_id DROP NOT NULL;
ALTER TABLE icy_transactions ALTER COLUMN src_employee_id DROP NOT NULL;

ALTER TABLE icy_transactions ADD COLUMN sender TEXT;
ALTER TABLE icy_transactions ADD COLUMN target TEXT;

ALTER TABLE icy_transactions DROP CONSTRAINT unique_icy_txn_src_dest_category;
ALTER TABLE icy_transactions ADD CONSTRAINT unique_icy_txn_src_dest_category UNIQUE (sender, target, category, amount, txn_time);

CREATE OR REPLACE VIEW "vw_icy_employee_dashboard" AS
SELECT
    t.dest_employee_id as employee_id,
    e.full_name,
    e.team_email,
    e.personal_email,
    SUM(t.amount) AS "total_earned"
FROM
    icy_transactions t
        JOIN employees e ON e.id = t.dest_employee_id
GROUP BY
    t.dest_employee_id,
    e.full_name,
    e.team_email,
    e.personal_email
ORDER BY
    SUM(t.amount) DESC;

-- +migrate Down
ALTER TABLE icy_transactions DROP CONSTRAINT unique_icy_txn_src_dest_category;
ALTER TABLE icy_transactions ADD CONSTRAINT unique_icy_txn_src_dest_category UNIQUE (src_employee_id, dest_employee_id, category, amount, txn_time);

ALTER TABLE icy_transactions DROP COLUMN sender;
ALTER TABLE icy_transactions DROP COLUMN target;

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
