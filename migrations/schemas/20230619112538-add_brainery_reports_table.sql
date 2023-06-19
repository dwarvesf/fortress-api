-- +migrate Up
CREATE TABLE IF NOT EXISTS “brainery_logs” (
    id              UUID PRIMARY KEY DEFAULT (UUID()),
    deleted_at      TIMESTAMP(6),
    created_at      TIMESTAMP(6) DEFAULT (now()),
    updated_at      TIMESTAMP(6) DEFAULT (now()),

    title           TEXT NOT NULL,
    url             TEXT NOT NULL,
    github_id       TEXT,
    discord_id      TEXT NOT NULL,
    employee_id     UUID DEFAULT NULL,
    tags            JSONB,
    published_at    TIMESTAMP(6) NOT NULL,
    reward          DECIMAL
);

ALTER TABLE brainery_logs
    ADD CONSTRAINT brainery_logs_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);

-- +migrate Down
DROP TABLE IF EXISTS brainery_logs;
