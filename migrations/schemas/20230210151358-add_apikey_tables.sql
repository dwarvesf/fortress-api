
-- +migrate Up

CREATE TABLE IF NOT EXISTS apikeys (
	id uuid PRIMARY KEY DEFAULT(uuid()),
	deleted_at TIMESTAMP(6),
	created_at TIMESTAMP(6) DEFAULT(NOW()),
	updated_at TIMESTAMP(6) DEFAULT(NOW()),
	client_id TEXT,
	secret_key TEXT,
	status TEXT
);

CREATE TABLE IF NOT EXISTS apikey_roles (
    id          uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at  TIMESTAMP(6),
    created_at  TIMESTAMP(6)     DEFAULT (NOW()),
    updated_at  TIMESTAMP(6)     DEFAULT (NOW()),
    apikey_id uuid NOT NULL,
    role_id uuid NOT NULL
);

ALTER TABLE apikey_roles
    ADD CONSTRAINT apikey_roles_apikey_id_fkey FOREIGN KEY (apikey_id) REFERENCES apikeys (id);

ALTER TABLE apikey_roles
    ADD CONSTRAINT apikey_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles (id);

-- +migrate Down

DROP TABLE IF EXISTS apikey_roles;

DROP TABLE IF EXISTS apikeys;