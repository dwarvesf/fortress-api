
-- +migrate Up
CREATE VIEW "vw_monthly_delivery_metrics" AS
SELECT 
	DATE_TRUNC('month', date) as "month",
	SUM(weight) AS "sum_weight",
	SUM(effort) AS "sum_effort"
FROM delivery_metrics
GROUP BY DATE_TRUNC('month', date)
ORDER BY "month" DESC;

-- +migrate Down
DROP VIEW "vw_monthly_delivery_metrics";
