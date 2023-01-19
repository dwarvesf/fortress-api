
-- +migrate Up
INSERT INTO permissions (id, deleted_at, created_at, updated_at, name, code) VALUES
('5c1745b6-d920-47d2-986a-fe6c48802ace', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read Line Manager', 'employees.read.lineManager.fullAccess');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, '5c1745b6-d920-47d2-986a-fe6c48802ace'
FROM roles r
WHERE r.code = 'admin' OR r.code = 'engineering-manager';

-- +migrate Down
DELETE FROM role_permissions 
WHERE permission_id = '5c1745b6-d920-47d2-986a-fe6c48802ace';

DELETE FROM public.permissions 
WHERE id = '5c1745b6-d920-47d2-986a-fe6c48802ace';
