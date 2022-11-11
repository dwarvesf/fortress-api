-- +migrate Up
CREATE TABLE IF NOT EXISTS seniorities (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (now()),
    updated_at TIMESTAMP(6)     DEFAULT (now()),

    name       TEXT,
    code       TEXT
);

ALTER TABLE employees ADD IF NOT EXISTS seniority_id uuid;
ALTER TABLE employees ADD CONSTRAINT employees_seniority_id_fkey FOREIGN KEY (seniority_id) REFERENCES seniorities (id);

-- +migrate Down
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_seniority_id_fkey;
ALTER TABLE employees DROP COLUMN IF EXISTS seniority_id;

DROP TABLE IF EXISTS seniorities;
