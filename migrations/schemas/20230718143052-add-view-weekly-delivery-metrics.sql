
-- +migrate Up
CREATE INDEX delivery_metrics_date_idx ON delivery_metrics ((date::date) DESC NULLS LAST);

CREATE VIEW "vw_weekly_delivery_metrics" AS
SELECT "date", 
	SUM(weight) AS "sum_weight", 
	SUM(effort) AS "sum_effort"
FROM delivery_metrics
GROUP BY "date";

-- +migrate Down
DROP VIEW "vw_weekly_delivery_metrics";

DROP INDEX IF EXISTS delivery_metrics_date_idx;
