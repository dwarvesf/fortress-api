-- +migrate Up
INSERT INTO public.permissions (id, deleted_at, created_at, updated_at, name, code) VALUES
('738ae6ae-e961-4fc2-b941-079e130f4213', NULL, '2023-02-13 13:32:29.330571', '2023-02-13 13:32:29.330571', 'Bank Account Read', 'bankAccounts.read'),
('75ae4f79-cef1-4dfd-be66-064c118ba6cf', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Clients Create', 'clients.create'),
('97a02a4d-b8fb-486d-af1a-25139014b375', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Clients Read', 'clients.read'),
('92031f1a-b745-42cf-aac1-a067bbee08e6', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Clients Edit', 'clients.edit'),
('7fb332e7-3231-42e6-8f38-a5c8b224d84e', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Clients Delete', 'clients.delete');

INSERT INTO public.role_permissions (role_id, permission_id)
SELECT r.id, '738ae6ae-e961-4fc2-b941-079e130f4213'
FROM roles r
WHERE r.code = 'admin' OR r.code = 'engineering-manager';

INSERT INTO public.role_permissions (role_id, permission_id)
SELECT r.id, '75ae4f79-cef1-4dfd-be66-064c118ba6cf'
FROM roles r
WHERE r.code = 'admin' OR r.code = 'engineering-manager';

INSERT INTO public.role_permissions (role_id, permission_id)
SELECT r.id, '97a02a4d-b8fb-486d-af1a-25139014b375'
FROM roles r
WHERE r.code = 'admin' OR r.code = 'engineering-manager';

INSERT INTO public.role_permissions (role_id, permission_id)
SELECT r.id, '92031f1a-b745-42cf-aac1-a067bbee08e6'
FROM roles r
WHERE r.code = 'admin' OR r.code = 'engineering-manager';

INSERT INTO public.role_permissions (role_id, permission_id)
SELECT r.id, '7fb332e7-3231-42e6-8f38-a5c8b224d84e'
FROM roles r
WHERE r.code = 'admin' OR r.code = 'engineering-manager';
-- +migrate Down
