
-- +migrate Up
INSERT INTO public.permissions (id, deleted_at, created_at, updated_at, name, code) VALUES
('738ae6ae-e961-4fc2-b941-079e130f4213', NULL, '2023-02-13 13:32:29.330571', '2023-02-13 13:32:29.330571', 'Bank Account Read', 'bankAccounts.read');

INSERT INTO public.role_permissions (role_id, permission_id)
SELECT r.id, '738ae6ae-e961-4fc2-b941-079e130f4213'
FROM roles r
WHERE r.code = 'admin' OR r.code = 'engineering-manager';

-- +migrate Down
