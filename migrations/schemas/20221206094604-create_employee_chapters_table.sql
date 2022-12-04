-- +migrate Up
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_chapter_id_fkey;
ALTER TABLE employees DROP COLUMN IF EXISTS chapter_id;

CREATE TABLE IF NOT EXISTS "employee_chapters" (
    "id"            uuid PRIMARY KEY DEFAULT (uuid()),
    "deleted_at"    TIMESTAMP(6),
    "created_at"    TIMESTAMP(6) DEFAULT (now()),
    "updated_at"    TIMESTAMP(6) DEFAULT (now()),
    "employee_id"   uuid NOT NULL,
    "chapter_id"      uuid NOT NULL
);

ALTER TABLE "employee_chapters" ADD CONSTRAINT employee_chapters_chapter_id_fkey FOREIGN KEY ("chapter_id") REFERENCES "chapters" ("id");

ALTER TABLE "employee_chapters" ADD CONSTRAINT employee_chapters_employee_id_fkey FOREIGN KEY ("employee_id") REFERENCES "employees" ("id");

ALTER TABLE chapters ADD IF NOT EXISTS lead_id uuid;
ALTER TABLE chapters ADD CONSTRAINT chapters_lead_id_fkey FOREIGN KEY (lead_id) REFERENCES employees (id);

-- +migrate Down
ALTER TABLE chapters DROP CONSTRAINT IF EXISTS chapters_lead_id_fkey;
ALTER TABLE chapters DROP COLUMN IF EXISTS lead_id;

ALTER TABLE "employee_chapters" DROP CONSTRAINT IF EXISTS employee_chapters_chapter_id_fkey;

ALTER TABLE "employee_chapters" DROP CONSTRAINT IF EXISTS employee_chapters_employee_id_fkey;

DROP TABLE IF EXISTS employee_chapters;

ALTER TABLE employees ADD IF NOT EXISTS chapter_id uuid;
ALTER TABLE employees ADD CONSTRAINT employees_chapter_id_fkey FOREIGN KEY (chapter_id) REFERENCES chapters (id);
