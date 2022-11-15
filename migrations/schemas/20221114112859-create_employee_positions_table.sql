
-- +migrate Up
CREATE TABLE IF NOT EXISTS employee_positions (
    id              uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at      TIMESTAMP(6),
    created_at      TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at      TIMESTAMP(6)     DEFAULT (NOW()),
    employee_id     uuid NOT NULL,
    position_id     uuid NOT NULL
);  

ALTER TABLE employees DROP COLUMN IF EXISTS possiton_id;

ALTER TABLE employee_positions
    ADD CONSTRAINT employee_positions_position_id_fkey FOREIGN KEY (position_id) REFERENCES positions(id);

ALTER TABLE employee_positions
    ADD CONSTRAINT employee_positions_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees(id);

-- +migrate Down
ALTER TABLE employee_positions DROP CONSTRAINT IF EXISTS employee_positions_position_id_fkey;

ALTER TABLE employee_positions DROP CONSTRAINT IF EXISTS employee_positions_employee_id_fkey;

ALTER TABLE employees ADD IF NOT EXISTS position_id uuid;

DROP TABLE IF EXISTS employee_positions;
