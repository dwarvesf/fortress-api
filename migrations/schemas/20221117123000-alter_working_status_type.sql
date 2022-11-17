
-- +migrate Up
ALTER TYPE working_status ADD VALUE IF NOT EXISTS  'on-boarding';

-- +migrate Down
