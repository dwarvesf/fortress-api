-- +migrate Up
CREATE TABLE IF NOT EXISTS "monthly_chapter_members" (
    id                 uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at         TIMESTAMP(6),
    created_at         TIMESTAMP(6)     DEFAULT (now()),
    updated_at         TIMESTAMP(6)     DEFAULT (now()),

    month              DATE,
    chapter_group_name TEXT NOT NULL,
    total_member       INT
);

CREATE TABLE IF NOT EXISTS "weekly_project_charge_rates" (
    id                 UUID PRIMARY KEY DEFAULT (UUID()),
    deleted_at         TIMESTAMP(6),
    created_at         TIMESTAMP(6)     DEFAULT (now()),
    updated_at         TIMESTAMP(6)     DEFAULT (now()),

    week               DATE,
    project_id         UUID NOT NULL,
    project_name       TEXT NOT NULL,
    member_id          UUID NOT NULL,
    member_name        TEXT NOT NULL,
    charge_rate_amount DECIMAL,
    deployment_type    deployment_types
);

ALTER TABLE weekly_project_charge_rates
    ADD CONSTRAINT weekly_project_charge_rates_member_id_fkey FOREIGN KEY (member_id) REFERENCES employees (id);

ALTER TABLE weekly_project_charge_rates
    ADD CONSTRAINT weekly_project_charge_rates_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);

-- +migrate Down
DROP TABLE IF EXISTS monthly_chapter_members;
DROP TABLE IF EXISTS weekly_project_charge_rates;
