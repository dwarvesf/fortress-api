
-- +migrate Up
DROP TYPE IF EXISTS "public"."cost_types";
CREATE TYPE "public"."cost_types" AS ENUM ('individual', 'sum');
DROP TYPE IF EXISTS "public"."amount_types";
CREATE TYPE "public"."amount_types" AS ENUM ('percentage', 'flat');

-- Table Definition
CREATE TABLE "public"."cost_projections" (
    "id" uuid NOT NULL DEFAULT uuid(),
    "name" varchar(255) NOT NULL,
    "type" "public"."cost_types" NOT NULL,
    "amount_type" "public"."amount_types" NOT NULL,
    "amount" float8 NOT NULL,
    PRIMARY KEY ("id")
);

CREATE OR REPLACE VIEW vw_project_cost_and_revenue AS
WITH project_data AS (
    -- Summarize project salary and charge rate
    SELECT
        ss.project_id,
        ss.project_name,
        SUM(ss.total_salary) AS total_salary,
        SUM(ss.charge_rate) AS total_estimated_profit,
        COUNT(DISTINCT ss.employee_id) AS total_project_members
    FROM vw_project_salary_summary ss
    GROUP BY ss.project_id, ss.project_name
),
company_data AS (
    -- Calculate the total number of employees in the company
    SELECT COUNT(DISTINCT id) AS total_company_employees
    FROM employees
),
latest_conversion_rates AS (
    -- Simulate getting the latest conversion rate for currencies (same logic as before)
    SELECT DISTINCT ON (currency_id)
        currency_id,
        to_usd,
        to_vnd
    FROM conversion_rates
    WHERE deleted_at IS NULL
    ORDER BY currency_id, created_at DESC
),
cost_projection_calculations AS (
    SELECT 
        p.project_id,
        SUM(
            CASE -- Convert VND cost projections to USD before applying calculations
                -- Individual, Flat amount (convert VND to USD first)
                WHEN cp.type = 'individual' AND cp.amount_type = 'flat' 
                THEN (cp.amount * COALESCE(lcr.to_usd, 1)) * p.total_project_members 

                -- Individual, Percentage amount (convert VND to USD first)
                WHEN cp.type = 'individual' AND cp.amount_type = 'percentage' 
                THEN (cp.amount * COALESCE(lcr.to_usd, 1)) * p.total_salary / 100
                
                -- Sum, Flat amount (converted from VND to USD)
                WHEN cp.type = 'sum' AND cp.amount_type = 'flat' 
                THEN (cp.amount * COALESCE(lcr.to_usd, 1)) / COALESCE(c.total_company_employees, 1) * p.total_project_members

                -- Sum, Percentage amount (converted from VND to USD)
                WHEN cp.type = 'sum' AND cp.amount_type = 'percentage' 
                THEN (cp.amount * COALESCE(lcr.to_usd, 1)) * p.total_salary / 100

                -- Default to 0 if no cost adjustments
                ELSE 0 
            END
        ) AS cost_projection_adjustment
    FROM project_data p
    CROSS JOIN company_data c
    LEFT JOIN cost_projections cp ON 1 = 1 -- Apply all cost projections
    -- Join on latest_conversion_rates to get the VND to USD conversion rate
    LEFT JOIN latest_conversion_rates lcr ON lcr.currency_id = (
        SELECT id FROM currencies WHERE name = 'VND' LIMIT 1
    )
    GROUP BY p.project_id
)
SELECT
    pd.project_name,

    -- Calculate total estimated cost as total project salary plus any cost projections
    -- and add 10% of the estimated profit to the estimated cost
    CEIL(pd.total_salary + COALESCE(cpc.cost_projection_adjustment, 0) + pd.total_estimated_profit * 0.1) AS estimated_cost,

    -- Total estimated profit comes directly from project data
    CEIL(pd.total_estimated_profit) AS estimated_revenue
FROM project_data pd
LEFT JOIN cost_projection_calculations cpc ON cpc.project_id = pd.project_id;


-- +migrate Down
DROP VIEW IF EXISTS "vw_project_cost_and_revenue";
DROP TABLE IF EXISTS "cost_projections";
