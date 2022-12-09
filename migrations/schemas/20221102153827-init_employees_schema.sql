-- +migrate Up
CREATE OR REPLACE FUNCTION uuid() RETURNS uuid
    LANGUAGE c
AS
    '$libdir/pgcrypto',
    'pg_random_uuid';

CREATE TYPE working_status AS ENUM (
    'left',
    'probation',
    'full-time',
    'contractor',
    'on-boarding'
);

CREATE TABLE IF NOT EXISTS stacks (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at TIMESTAMP(6)     DEFAULT (NOW()),
    name       TEXT,
    code       TEXT,
    avatar     TEXT
);

CREATE TABLE IF NOT EXISTS countries (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (now()),
    updated_at TIMESTAMP(6)     DEFAULT (now()),
    name       TEXT,
    code       TEXT,
    cities     JSONB            DEFAULT '[]'::JSONB
);

CREATE TABLE IF NOT EXISTS positions (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (now()),
    updated_at TIMESTAMP(6)     DEFAULT (now()),
    name       TEXT,
    code       TEXT
);

CREATE TABLE IF NOT EXISTS seniorities (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (now()),
    updated_at TIMESTAMP(6)     DEFAULT (now()),

    name       TEXT,
    code       TEXT
);

CREATE TABLE IF NOT EXISTS employees (
    id                        uuid PRIMARY KEY DEFAULT uuid(),
    deleted_at                TIMESTAMP(6),
    created_at                TIMESTAMP(6)     DEFAULT NOW(),
    updated_at                TIMESTAMP(6)     DEFAULT NOW(),
    full_name                 TEXT NOT NULL,
    display_name              TEXT NOT NULL,
    gender                    TEXT NOT NULL,
    team_email                TEXT NOT NULL,
    personal_email            TEXT NOT NULL,
    avatar                    TEXT NOT NULL,
    phone_number              TEXT,
    address                   TEXT,
    mbti                      TEXT,
    horoscope                 TEXT,
    passport_photo_front      TEXT,
    passport_photo_back       TEXT,
    identity_card_photo_front TEXT,
    identity_card_photo_back  TEXT,
    date_of_birth             DATE,
    working_status            working_status,
    joined_date               DATE,
    left_date                 DATE,
    basecamp_id               TEXT,
    basecamp_attachable_sgid  TEXT,
    gitlab_id                 TEXT,
    github_id                 TEXT,
    discord_id                TEXT,
    notion_id                 TEXT,
    wise_recipient_email      TEXT,
    wise_recipient_name       TEXT,
    wise_recipient_id         TEXT,
    wise_account_number       TEXT,
    wise_currency             TEXT,
    local_bank_branch         TEXT,
    local_bank_number         TEXT,
    local_bank_currency       TEXT,
    local_branch_name         TEXT,
    local_bank_recipient_name TEXT,
    seniority_id              uuid,
    line_manager_id           uuid
);

CREATE TABLE IF NOT EXISTS employee_positions (
    id          uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at  TIMESTAMP(6),
    created_at  TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at  TIMESTAMP(6)     DEFAULT (NOW()),
    employee_id uuid NOT NULL,
    position_id uuid NOT NULL
);

CREATE TABLE employee_stacks (
    id              uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at      TIMESTAMP(6),
    created_at      TIMESTAMP(6)     DEFAULT (now()),
    updated_at      TIMESTAMP(6)     DEFAULT (now()),
    employee_id     uuid NOT NULL,
    stack_id        uuid NOT NULL
);

CREATE TABLE IF NOT EXISTS chapters (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (now()),
    updated_at TIMESTAMP(6)     DEFAULT (now()),
    name       TEXT,
    code       TEXT,
    lead_id    uuid
);

CREATE TABLE IF NOT EXISTS employee_chapters (
    id          uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at  TIMESTAMP(6),
    created_at  TIMESTAMP(6)     DEFAULT (now()),
    updated_at  TIMESTAMP(6)     DEFAULT (now()),
    employee_id uuid NOT NULL,
    chapter_id  uuid NOT NULL
);

ALTER TABLE employees
    ADD CONSTRAINT employees_seniority_id_fkey FOREIGN KEY (seniority_id) REFERENCES seniorities (id);

ALTER TABLE employees
    ADD CONSTRAINT employee_line_manager_id_fkey FOREIGN KEY (line_manager_id) REFERENCES employees (id);

ALTER TABLE employee_positions
    ADD CONSTRAINT employee_positions_position_id_fkey FOREIGN KEY (position_id) REFERENCES positions (id);

ALTER TABLE employee_positions
    ADD CONSTRAINT employee_positions_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);

ALTER TABLE employee_stacks
    ADD CONSTRAINT employee_stacks_stack_id_fkey FOREIGN KEY (stack_id) REFERENCES stacks (id);

ALTER TABLE employee_stacks
    ADD CONSTRAINT employee_stacks_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);

ALTER TABLE employee_chapters
    ADD CONSTRAINT employee_chapters_chapter_id_fkey FOREIGN KEY (chapter_id) REFERENCES chapters (id);

ALTER TABLE employee_chapters
    ADD CONSTRAINT employee_chapters_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);

ALTER TABLE chapters
    ADD CONSTRAINT chapters_lead_id_fkey FOREIGN KEY (lead_id) REFERENCES employees (id);

-- +migrate Down
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_seniority_id_fkey;
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_chapter_id_fkey;
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employee_line_manager_id_fkey;
ALTER TABLE employee_positions DROP CONSTRAINT IF EXISTS employee_positions_position_id_fkey;
ALTER TABLE employee_positions DROP CONSTRAINT IF EXISTS employee_positions_employee_id_fkey;
ALTER TABLE employee_stacks DROP CONSTRAINT IF EXISTS employee_stacks_stack_id_fkey;
ALTER TABLE employee_stacks DROP CONSTRAINT IF EXISTS employee_stacks_employee_id_fkey;
ALTER TABLE employee_chapters DROP CONSTRAINT IF EXISTS employee_chapters_chapter_id_fkey;
ALTER TABLE employee_chapters DROP CONSTRAINT IF EXISTS employee_chapters_employee_id_fkey;
ALTER TABLE chapters DROP CONSTRAINT IF EXISTS chapters_lead_id_fkey;

DROP TABLE IF EXISTS employee_chapters;
DROP TABLE IF EXISTS chapters;
DROP TABLE IF EXISTS employee_stacks;
DROP TABLE IF EXISTS employee_positions;
DROP TABLE IF EXISTS seniorities;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS stacks;
DROP TABLE IF EXISTS countries;
DROP TABLE IF EXISTS employees;

DROP TYPE IF EXISTS working_status;
