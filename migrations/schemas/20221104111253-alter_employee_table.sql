
-- +migrate Up
alter type employment_status rename to working_status;
alter table employees
    rename column employment_status to "working_status";
-- +migrate Down
alter type working_status rename to employment_status;
alter table employees
    rename column working_status to "employment_status";
