-- +migrate Up
ALTER TYPE invoice_statuses ADD VALUE IF NOT EXISTS 'uncollectible';

-- +migrate Down
-- Note: PostgreSQL does not support removing enum values directly
-- The 'uncollectible' status will remain in the enum type
