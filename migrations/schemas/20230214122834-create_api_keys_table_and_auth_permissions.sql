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


INSERT INTO "public"."permissions" ("deleted_at", "created_at", "updated_at", "name", "code") VALUES
(NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Auth Create', 'auth.create');

INSERT INTO role_permissions ("role_id", "permission_id")
SELECT
	roles.id,
	permissions.id
FROM
	roles,
	permissions
WHERE
	roles.code = 'admin'
	AND permissions.code = 'auth.create';
-- +migrate Down
DROP TABLE IF EXISTS api_key_roles;
DROP TABLE IF EXISTS api_keys;

DELETE FROM role_permissions
WHERE permission_id = (
		SELECT
			id
		FROM
			permissions
		WHERE
			code = 'auth.create');

DELETE FROM permissions WHERE code='auth.create';
