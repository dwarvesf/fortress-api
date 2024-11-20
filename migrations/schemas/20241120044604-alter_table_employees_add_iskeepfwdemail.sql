
-- +migrate Up
ALTER TABLE employees ADD COLUMN is_keep_fwd_email BOOLEAN DEFAULT FALSE;

-- +migrate Down
ALTER TABLE employees DROP COLUMN is_keep_fwd_email;
