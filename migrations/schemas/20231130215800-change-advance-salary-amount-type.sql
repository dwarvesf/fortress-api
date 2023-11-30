-- +migrate Up
ALTER TABLE payrolls
ALTER COLUMN salary_advance_amount TYPE FLOAT8 USING salary_advance_amount::FLOAT8;

ALTER TABLE payrolls
    ALTER COLUMN salary_advance_amount SET DEFAULT 0;

-- +migrate Down

ALTER TABLE payrolls
ALTER COLUMN salary_advance_amount TYPE INT8 USING salary_advance_amount::INT8;

ALTER TABLE payrolls
    ALTER COLUMN salary_advance_amount SET DEFAULT NULL;