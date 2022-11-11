-- +migrate Up
CREATE TABLE IF NOT EXISTS chapters (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (now()),
    updated_at TIMESTAMP(6)     DEFAULT (now()),
    name       TEXT,
    code       TEXT
);
ALTER TABLE employees ADD IF NOT EXISTS chapter_id uuid;
ALTER TABLE employees ADD CONSTRAINT employees_chapter_id_fkey FOREIGN KEY (chapter_id) REFERENCES chapters (id);
-- +migrate Down

ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_chapter_id_fkey;
ALTER TABLE employees DROP COLUMN IF EXISTS chapter_id;

DROP TABLE IF EXISTS chapters;
