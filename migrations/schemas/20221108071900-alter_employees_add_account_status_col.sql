-- +migrate Up
CREATE TYPE account_status as ENUM('on-boarding', 'probation', 'active', 'on-leave');

ALTER TABLE employees ADD account_status account_status;

-- +migrate Down
DROP TYPE account_status;

ALTER TABLE employees DROP COLUMN IF EXISTS account_status;
