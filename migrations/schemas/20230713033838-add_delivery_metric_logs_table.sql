-- +migrate Up
CREATE TABLE IF NOT EXISTS delivery_metrics (
    id            UUID PRIMARY KEY DEFAULT (UUID()),
    deleted_at    TIMESTAMP(6),
    created_at    TIMESTAMP(6)     DEFAULT (now()),
    updated_at    TIMESTAMP(6)     DEFAULT (now()),

    weight        DECIMAL NOT NULL DEFAULT 0,
    effort        DECIMAL NOT NULL DEFAULT 0,
    effectiveness DECIMAL NOT NULL DEFAULT 0,
    date          DATE    NOT NULL,
    employee_id   UUID    NOT NULL,
    project_id    UUID    NOT NULL
);

ALTER TABLE delivery_metrics
    ADD CONSTRAINT delivery_metrics_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);

ALTER TABLE delivery_metrics
    ADD CONSTRAINT delivery_metrics_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);

-- +migrate Down
DROP TABLE IF EXISTS delivery_metrics;
