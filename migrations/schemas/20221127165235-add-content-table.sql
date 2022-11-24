-- +migrate Up
CREATE TABLE IF NOT EXISTS contents (
    id          uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at  TIMESTAMP(6),
    created_at  TIMESTAMP(6)     DEFAULT (now()),
    updated_at  TIMESTAMP(6)     DEFAULT (now()),

    type        TEXT,
    extension   TEXT,
    path        TEXT,
    upload_by   uuid,
    employee_id uuid
);

ALTER TABLE contents ADD CONSTRAINT contents_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);
ALTER TABLE contents ADD CONSTRAINT contents_upload_by_fkey FOREIGN KEY (upload_by) REFERENCES employees (id);

-- +migrate Down
ALTER TABLE contents DROP CONSTRAINT contents_employee_id_fkey;
ALTER TABLE contents DROP CONSTRAINT contents_upload_by_fkey;
DROP TABLE IF EXISTS contents;

