-- +migrate Up
CREATE TABLE IF NOT EXISTS positions (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (now()),
    updated_at TIMESTAMP(6)     DEFAULT (now()),
    name       TEXT,
    code       TEXT
);

ALTER TABLE employees ADD position_id uuid;
ALTER TABLE employees ADD CONSTRAINT employees_position_id_fkey FOREIGN KEY (position_id) REFERENCES positions (id);

-- +migrate Down
ALTER TABLE employees DROP CONSTRAINT employees_position_id_fkey;
ALTER TABLE employees DROP COLUMN position_id;

DROP TABLE IF EXISTS positions;
