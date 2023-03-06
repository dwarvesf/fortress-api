
-- +migrate Up
alter table audits
    alter column auditor_id drop not null;

-- +migrate Down
alter table audits
    alter column auditor_id set not null;
