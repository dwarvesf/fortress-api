INSERT INTO public.positions (id, deleted_at, created_at, updated_at, name, code) VALUES
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Frontend', 'frontend'),
('d796884d-a8c4-4525-81e7-54a3b6099eac', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Backend', 'backend'),
('dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Devops', 'devops'),
('01fb6322-d727-47e3-a242-5039ea4732fc', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Blockchain', 'blockchain'),
('39735742-829b-47f3-8f9d-daf0983914e5', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Project-Management', 'project-management');

INSERT INTO public.roles (id, deleted_at, created_at, updated_at, name, code) VALUES
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Admin', 'admin'),
('d796884d-a8c4-4525-81e7-54a3b6099eac', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Member', 'member');

INSERT INTO public.seniorities (id, deleted_at, created_at, updated_at, name, code) VALUES
('01fb6322-d727-47e3-a242-5039ea4732fd', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Staff', 'staff'),
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Frontend', 'frontend');

INSERT INTO public.chapters (id, deleted_at, created_at, updated_at, name, code) VALUES
('39735742-829b-47f3-8f9d-daf0983914e5', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'PM', 'pm');

INSERT INTO public.employees (id,deleted_at,created_at,updated_at,full_name,display_name,gender,team_email,personal_email,avatar,phone_number,address,mbti,horoscope,passport_photo_front,passport_photo_back,identity_card_photo_front,identity_card_photo_back,date_of_birth,"working_status",joined_date,left_date,basecamp_id,basecamp_attachable_sgid,gitlab_id,github_id,discord_id,wise_recipient_email,wise_recipient_name,wise_recipient_id,wise_account_number,wise_currency,local_bank_branch,local_bank_number,local_bank_currency,local_branch_name,local_bank_recipient_name,seniority_id,notion_id,line_manager_id) VALUES
('2655832e-f009-4b73-a535-64c3a22e558f', NULL, '2022-11-02 09:52:34.586566', '2022-11-02 09:52:34.586566', 'Phạm Đức Thành','Thanh Pham','Male','thanh@d.foundation','thanhpham123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/5153574695663955944.png','0123456788','Tan Binh District, Ho Chi Minh, Vietnam','ISFJ-A','Aquaman',NULL,NULL,NULL,NULL,'1990-01-02','contractor','2018-09-01',NULL,NULL,NULL,'thanhpham','thanhpham','646649040771219476',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'01fb6322-d727-47e3-a242-5039ea4732fd','39523061',NULL);

INSERT INTO public.employee_roles (id, deleted_at, created_at, updated_at, employee_id, role_id) VALUES
('b3c27435-eb39-400b-a94b-7bf8218b8b3e', NULL, '2022-11-11 18:36:31.657696', '2022-11-11 18:36:31.657696', '2655832e-f009-4b73-a535-64c3a22e558f', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8');

