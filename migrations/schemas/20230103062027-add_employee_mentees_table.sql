
-- +migrate Up
CREATE TABLE employee_mentees (
    id              uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at      TIMESTAMP(6),
    created_at      TIMESTAMP(6)     DEFAULT (now()),
    updated_at      TIMESTAMP(6)     DEFAULT (now()),
    mentor_id       uuid NOT NULL,
    mentee_id       uuid NOT NULL
);

ALTER TABLE employee_mentees
    ADD CONSTRAINT employee_mentees_mentor_id_fkey FOREIGN KEY (mentor_id) REFERENCES employees (id);

ALTER TABLE employee_mentees
    ADD CONSTRAINT employee_mentees_mentee_id_fkey FOREIGN KEY (mentee_id) REFERENCES employees (id);

-- +migrate Down
ALTER TABLE employee_mentees DROP CONSTRAINT IF EXISTS employee_mentees_mentee_id_fkey;
ALTER TABLE employee_mentees DROP CONSTRAINT IF EXISTS employee_mentees_mentor_id_fkey;
DROP TABLE IF EXISTS employee_mentees;
