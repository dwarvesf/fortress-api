-- +migrate Up
CREATE TYPE account_status as ENUM('on-boarding', 'probation', 'active', 'on-leave');

ALTER TABLE employees ADD IF NOT EXISTS account_status account_status;

-- +migrate Down
ALTER TABLE employees DROP COLUMN IF EXISTS account_status;

DROP TYPE IF EXISTS account_status;
