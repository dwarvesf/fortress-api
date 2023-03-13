
-- +migrate Up
ALTER TABLE project_members ADD COLUMN upsell_commission_rate DECIMAL;

-- +migrate Down
ALTER TABLE project_members DROP COLUMN IF EXISTS upsell_commission_rate;
