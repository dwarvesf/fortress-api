-- Add 'deal-closing' to the project_head_positions ENUM type
-- +migrate Up
ALTER TYPE project_head_positions ADD VALUE IF NOT EXISTS 'deal-closing';

-- +migrate Down
-- No-op: removing ENUM values is not supported in PostgreSQL
SELECT
    1;
