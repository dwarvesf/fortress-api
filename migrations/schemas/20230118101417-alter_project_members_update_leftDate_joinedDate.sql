
-- +migrate Up
ALTER TABLE project_members RENAME COLUMN joined_date TO start_date;
ALTER TABLE project_members RENAME COLUMN left_date TO end_date;

ALTER TABLE project_heads RENAME COLUMN joined_date TO start_date;
ALTER TABLE project_heads RENAME COLUMN left_date TO end_date;

ALTER TABLE work_unit_members RENAME COLUMN joined_date TO start_date;
ALTER TABLE work_unit_members RENAME COLUMN left_date TO end_date;

-- +migrate Down
ALTER TABLE project_members RENAME COLUMN end_date TO left_date;
ALTER TABLE project_members RENAME COLUMN start_date TO joined_date;

ALTER TABLE project_heads RENAME COLUMN end_date TO left_date;
ALTER TABLE project_heads RENAME COLUMN start_date TO joined_date;

ALTER TABLE work_unit_members RENAME COLUMN end_date TO left_date;
ALTER TABLE work_unit_members RENAME COLUMN start_date TO joined_date;
