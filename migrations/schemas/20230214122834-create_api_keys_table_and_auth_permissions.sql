-- +migrate Up

CREATE TABLE IF NOT EXISTS api_keys (
    id uuid PRIMARY KEY DEFAULT(uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6) DEFAULT(NOW()),
    updated_at TIMESTAMP(6) DEFAULT(NOW()),

    client_id TEXT,
    secret_key TEXT,
    status TEXT
    );

CREATE TABLE IF NOT EXISTS api_key_roles (
    id          uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at  TIMESTAMP(6),
    created_at  TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at  TIMESTAMP(6)     DEFAULT (NOW()),
    api_key_id uuid NOT NULL,
    role_id uuid NOT NULL
    );

ALTER TABLE api_key_roles
    ADD CONSTRAINT api_key_roles_api_key_id_fkey FOREIGN KEY (api_key_id) REFERENCES api_keys (id);

ALTER TABLE api_key_roles
    ADD CONSTRAINT api_key_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles (id);
-- +migrate Down
DROP TABLE IF EXISTS api_key_roles;
DROP TABLE IF EXISTS api_keys;