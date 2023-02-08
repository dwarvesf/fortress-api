-- +migrate Up
CREATE TABLE IF NOT EXISTS clients (
    id                  uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at          TIMESTAMP(6),
    created_at          TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at          TIMESTAMP(6)     DEFAULT (NOW()),

    name                TEXT,
    description         TEXT,
    registration_number TEXT,
    address             TEXT,
    country             TEXT,
    industry            TEXT,
    website             TEXT,
    emails              TEXT[]
);

CREATE TABLE IF NOT EXISTS client_contacts (
    id              uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at      TIMESTAMP(6),
    created_at      TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at      TIMESTAMP(6)     DEFAULT (NOW()),

    name            TEXT,
    client_id       UUID,
    role            TEXT,
    metadata        JSONB,
    emails          JSONB,
    is_main_contact BOOLEAN          DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS company_infos (
    id                  uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at          TIMESTAMP(6),
    created_at          TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at          TIMESTAMP(6)     DEFAULT (NOW()),

    name                TEXT,
    description         TEXT,
    registration_number TEXT,
    info                JSONB
);

CREATE TABLE IF NOT EXISTS invoice_number_caching (
    id         uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at TIMESTAMP(6)     DEFAULT (NOW()),

    key        TEXT,
    number     INT
);

ALTER TABLE projects ADD COLUMN "client_id" UUID NULL;
ALTER TABLE projects ADD COLUMN "company_info_id" UUID NULL;

ALTER TABLE client_contacts
    ADD CONSTRAINT client_contacts_client_id_fkey FOREIGN KEY (client_id) REFERENCES clients (id);
ALTER TABLE projects
    ADD CONSTRAINT projects_client_id_fkey FOREIGN KEY (client_id) REFERENCES clients (id);
ALTER TABLE projects
    ADD CONSTRAINT projects_company_info_id_fkey FOREIGN KEY (company_info_id) REFERENCES company_infos (id);

-- +migrate Down

ALTER TABLE client_contacts
    DROP CONSTRAINT IF EXISTS client_contacts_client_id_fkey;
ALTER TABLE projects
    DROP CONSTRAINT IF EXISTS projects_client_id_fkey;
ALTER TABLE projects DROP COLUMN client_id;
ALTER TABLE projects DROP COLUMN company_info_id;

DROP TABLE IF EXISTS clients;
DROP TABLE IF EXISTS client_contacts;
DROP TABLE IF EXISTS company_infos;
DROP TABLE IF EXISTS invoice_number_caching;
