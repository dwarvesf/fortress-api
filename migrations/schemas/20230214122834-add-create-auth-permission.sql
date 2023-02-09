
-- +migrate Up
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

DELETE FROM role_permissions
WHERE permission_id = (
		SELECT
			id
		FROM
			permissions
		WHERE
			code = 'auth.create');

DELETE FROM permissions WHERE code='auth.create';
