
-- +migrate Up
ALTER TABLE employees ALTER COLUMN working_status TYPE TEXT;

DROP TYPE IF EXISTS working_status;

CREATE TYPE working_status AS ENUM (
    'left', 
    'probation', 
    'full-time', 
    'contractor', 
    'on-boarding'
);

ALTER TABLE employees ALTER COLUMN working_status TYPE working_status USING(working_status::working_status);

-- +migrate Down
ALTER TABLE employees ALTER COLUMN working_status TYPE TEXT;

DROP TYPE IF EXISTS working_status;

CREATE TYPE working_status AS ENUM (
    'left', 
    'probation', 
    'full-time', 
    'contractor'
);

ALTER TABLE employees ALTER COLUMN working_status TYPE working_status USING(working_status::working_status);
