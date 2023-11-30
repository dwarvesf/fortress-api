-- +migrate Up
ALTER TABLE payrolls ADD COLUMN salary_advance_amount INT8;

-- +migrate Down
ALTER TABLE payrolls DROP COLUMN salary_advance_amount;
