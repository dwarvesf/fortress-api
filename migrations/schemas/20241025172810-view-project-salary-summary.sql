
-- +migrate Up
CREATE OR REPLACE VIEW vw_project_salary_summary AS
WITH latest_invoices AS (
    SELECT 
        inv.project_id,
        inv.total AS latest_invoice_total
    FROM invoices inv
    INNER JOIN (
        SELECT 
            project_id,
            MAX(created_at) AS latest_invoice_date
        FROM invoices
        GROUP BY project_id
    ) latest_invoice_info
    ON inv.project_id = latest_invoice_info.project_id
    AND inv.created_at = latest_invoice_info.latest_invoice_date
),
project_fixed_costs AS (
    SELECT 
        pm.project_id,
        li.latest_invoice_total AS total_invoice_amount,
        COUNT(DISTINCT pm.employee_id) AS total_members
    FROM project_members pm
    LEFT JOIN latest_invoices li ON li.project_id = pm.project_id
    WHERE pm.status = 'active'
    GROUP BY pm.project_id, li.latest_invoice_total
),
currency_usd AS (
    SELECT id AS usd_currency_id 
    FROM currencies 
    WHERE name = 'USD'
    LIMIT 1
),
latest_conversion_rates AS (
    SELECT DISTINCT ON (currency_id)
        currency_id,
        to_usd,
        to_vnd
    FROM conversion_rates
    WHERE deleted_at IS NULL
    ORDER BY currency_id, created_at DESC
)
SELECT DISTINCT
    pm.project_id,
    p.name AS project_name,

    CASE 
        WHEN ba_currency.name != 'USD' THEN 
            (CASE 
                WHEN p.type = 'time-material' THEN pm.rate
                WHEN p.type = 'dwarves' THEN 0
                WHEN p.type = 'fixed-cost' THEN COALESCE(pfc.total_invoice_amount / NULLIF(pfc.total_members, 0), 0)
                ELSE pm.rate 
            END) * COALESCE(ba_conv.to_usd, 1)
        ELSE 
            CASE 
                WHEN p.type = 'time-material' THEN pm.rate
                WHEN p.type = 'dwarves' THEN 0
                WHEN p.type = 'fixed-cost' THEN COALESCE(pfc.total_invoice_amount / NULLIF(pfc.total_members, 0), 0)
                ELSE pm.rate 
            END
    END AS charge_rate,

    -- Use project's currency via bank_account
    ba.currency_id,
    
    -- Convert salary to USD if necessary
    CASE 
        WHEN bs_curr.name != 'USD' THEN 
            (bs.company_account_amount + bs.personal_account_amount) * COALESCE(bs_conv.to_usd, 1)
        ELSE 
            (bs.company_account_amount + bs.personal_account_amount)
    END AS total_salary,
    
    pm.employee_id,
    pm.deployment_type,
    e.full_name AS employee_full_name
FROM project_members pm
LEFT JOIN projects p ON p.id = pm.project_id
LEFT JOIN employees e ON e.id = pm.employee_id
LEFT JOIN base_salaries bs ON bs.employee_id = pm.employee_id
    AND bs.is_active = true
    AND bs.deleted_at IS NULL
LEFT JOIN currencies bs_curr ON bs_curr.id = bs.currency_id
LEFT JOIN latest_conversion_rates bs_conv ON bs_conv.currency_id = bs.currency_id
LEFT JOIN project_fixed_costs pfc ON pfc.project_id = pm.project_id
LEFT JOIN bank_accounts ba ON ba.id = p.bank_account_id
LEFT JOIN currencies ba_currency ON ba.currency_id = ba_currency.id
LEFT JOIN latest_conversion_rates ba_conv ON ba_conv.currency_id = ba.currency_id
CROSS JOIN currency_usd cu
WHERE pm.status = 'active';

-- +migrate Down
DROP VIEW IF EXISTS "vw_project_salary_summary";