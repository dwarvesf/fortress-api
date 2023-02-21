
-- +migrate Up
ALTER TABLE contents DROP CONSTRAINT contents_employee_id_fkey;
ALTER TABLE contents DROP CONSTRAINT contents_upload_by_fkey;

ALTER TABLE contents RENAME COLUMN employee_id TO target_id;

ALTER TABLE contents ADD COLUMN target_type VARCHAR(20);

ALTER TABLE contents ADD COLUMN auth_type VARCHAR(20);


-- +migrate Down

ALTER TABLE contents RENAME COLUMN target_id TO employee_id;

ALTER TABLE contents ADD CONSTRAINT contents_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);
ALTER TABLE contents ADD CONSTRAINT contents_upload_by_fkey FOREIGN KEY (upload_by) REFERENCES employees (id);

ALTER TABLE contents DROP COLUMN target_type;

ALTER TABLE contents DROP COLUMN auth_type;

