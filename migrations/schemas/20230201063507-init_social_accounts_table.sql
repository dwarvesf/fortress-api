
-- +migrate Up
CREATE TABLE IF NOT EXISTS social_accounts (
    id              uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at      TIMESTAMP(6),
    created_at      TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at      TIMESTAMP(6)     DEFAULT (NOW()),
    employee_id     uuid,
    type            TEXT,
    account_id      TEXT,
    email           TEXT,
    display_name    TEXT
);

ALTER TABLE social_accounts
    ADD CONSTRAINT social_accounts_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);

INSERT INTO social_accounts (deleted_at, created_at, updated_at, employee_id, "type", account_id, email, display_name)
SELECT NULL, NOW(), NOW(), e.id, 'github', e.github_id, NULL, e.github_id
FROM employees e;

INSERT INTO social_accounts (deleted_at, created_at, updated_at, employee_id, "type", account_id, email, display_name)
SELECT NULL, NOW(), NOW(), e.id, 'gitlab', e.gitlab_id, NULL, e.gitlab_id 
FROM employees e;

INSERT INTO social_accounts (deleted_at, created_at, updated_at, employee_id, "type", account_id, email, display_name)
SELECT NULL, NOW(), NOW(), e.id, 'notion', e.notion_id, e.notion_email, e.notion_name 
FROM employees e;

INSERT INTO social_accounts (deleted_at, created_at, updated_at, employee_id, "type", account_id, email, display_name)
SELECT NULL, NOW(), NOW(), e.id, 'discord', e.discord_id, NULL, e.discord_name
FROM employees e;

INSERT INTO social_accounts (deleted_at, created_at, updated_at, employee_id, "type", account_id, email, display_name)
SELECT NULL, NOW(), NOW(), e.id, 'linkedin', e.linkedin_name, NULL, e.linkedin_name
FROM employees e;

-- +migrate Down
ALTER TABLE social_accounts DROP CONSTRAINT IF EXISTS social_accounts_employee_id_fkey;

DROP TABLE IF EXISTS social_accounts;
