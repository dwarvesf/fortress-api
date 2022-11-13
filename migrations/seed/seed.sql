INSERT INTO public.employees (id, deleted_at, created_at, updated_at, full_name, display_name, gender, team_email, personal_email, avatar, phone_number, address, mbti, horoscope, passport_photo_front, passport_photo_back, identity_card_photo_front, identity_card_photo_back, date_of_birth, working_status, joined_date, left_date, basecamp_id, basecamp_attachable_sgid, gitlab_id, github_id, discord_id, wise_recipient_email, wise_recipient_name, wise_recipient_id, wise_account_number, wise_currency, local_bank_branch, local_bank_number, local_bank_currency, local_branch_name, local_bank_recipient_name, position_id, seniority_id, chapter_id, account_status) VALUES
('ecea9d15-05ba-4a4e-9787-54210e3b98ce', null, '2022-11-02 09:50:55.320669', '2022-11-02 09:50:55.320669', 'Nguyễn Hoàng Huy', 'Huy Nguyen', 'Male', 'huynh@d.foundation', 'hoanghuy123@gmail.com', 'https://s3-ap-southeast-1.amazonaws.com/fortress-images/2830497479497502617.png', '0123456789', 'Somewhere his belong, Thu Duc City, Ho Chi Minh City', 'Defender', 'Virgo', null, null, null, null, '1990-01-02', 'probation', null, null, null, null, null, 'huynguyen', null, null, null, null, null, null, null, null, null, null, null, null, null, null, null),
('2655832e-f009-4b73-a535-64c3a22e558f', null, '2022-11-02 09:52:34.586566', '2022-11-02 09:52:34.586566', 'Phạm Đức Thành', 'Thanh Pham', 'Male', 'thanh@d.foundation', 'thanhpham123@gmail.com', 'https://s3-ap-southeast-1.amazonaws.com/fortress-images/5153574695663955944.png', '0123456788', 'Somewhere his belong, Tan Binh District, Ho Chi Minh, Vietnam', 'INTJ-A', 'Libra', null, null, null, null, '1990-01-02', 'contractor', null, null, null, null, null, 'thanhpham', null, null, null, null, null, null, null, null, null, null, null, null, null, null, null),
('8d7c99c0-3253-4286-93a9-e7554cb327ef', null, '2022-11-02 09:50:55.320669', '2022-11-02 09:50:55.320669', 'Nguyễn Hải Nam', 'Nam Nguyen', 'Male', 'benjamin@d.foundation', 'nam123@gmail.com', 'https://s3-ap-southeast-1.amazonaws.com/fortress-images/2830497479497502617.png', '0123456789', 'Somewhere his belong, District 7, Ho Chi Minh City', 'INTJ-S', 'Aquaman', null, null, null, null, '1990-01-02', 'probation', null, null, null, null, null, 'namnh', null, null, null, null, null, null, null, null, null, null, null, null, null, null, null),
('eeae589a-94e3-49ac-a94c-fcfb084152b2', null, '2022-11-02 09:50:55.320669', '2022-11-02 09:50:55.320669', 'Nguyễn Ngô Lập', 'Lap Nguyen', 'Male', 'alan@d.foundation', 'lap123@gmail.com', 'https://s3-ap-southeast-1.amazonaws.com/fortress-images/2830497479497502617.png', '0123456789', 'Somewhere his belong, District 7, Ho Chi Minh City', 'INTJ-S', 'Aquaman', null, null, null, null, '1990-01-02', 'probation', null, null, null, null, null, 'lapnn', null, null, null, null, null, null, null, null, null, null, null, null, null, null, null),
('608ea227-45a5-4c8a-af43-6c7280d96340', null, '2022-11-02 09:50:55.320669', '2022-11-02 09:50:55.320669', 'Tiêu Quang Huy', 'Huy Tieu', 'Male', 'huytq@d.foundation', 'huytq123@gmail.com', 'https://s3-ap-southeast-1.amazonaws.com/fortress-images/2830497479497502617.png', '0123456789', 'Somewhere his belong, District 7, Ho Chi Minh City', 'INTJ-S', 'Aquaman', null, null, null, null, '1990-01-02', 'contractor', null, null, null, null, null, 'huytq', null, null, null, null, null, null, null, null, null, null, null, null, null, null, null);


INSERT INTO public.roles (id, deleted_at, created_at, updated_at, name, code) VALUES
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Admin', 'admin'),
('d796884d-a8c4-4525-81e7-54a3b6099eac', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Member', 'member');

INSERT INTO public.employee_roles (id, deleted_at, created_at, updated_at, employee_id, role_id) VALUES
('f40ead54-dc57-40d5-9c6a-4f22341ee7d2', null, '2022-11-11 18:36:31.657696', '2022-11-11 18:36:31.657696', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('3fd72903-cebb-4519-82c2-4c8f096a272b', null, '2022-11-11 18:36:31.657696', '2022-11-11 18:36:31.657696', 'eeae589a-94e3-49ac-a94c-fcfb084152b2', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('b3c27435-eb39-400b-a94b-7bf8218b8b3e', null, '2022-11-11 18:36:31.657696', '2022-11-11 18:36:31.657696', '2655832e-f009-4b73-a535-64c3a22e558f', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('5bfea659-4912-4668-ae28-644083595aa6', null, '2022-11-11 18:36:31.657696', '2022-11-11 18:36:31.657696', '608ea227-45a5-4c8a-af43-6c7280d96340', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('178a7b5c-ddac-445d-9045-09d3b6b04431', null, '2022-11-11 18:36:31.657696', '2022-11-11 18:36:31.657696', '8d7c99c0-3253-4286-93a9-e7554cb327ef', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8');

INSERT INTO public.permissions (id, deleted_at, created_at, updated_at, name, code) VALUES
('01ed1076-4028-4fdf-9a92-cb57a8e041af', null, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read', 'employees.read'),
('a7549f30-987c-47d1-9266-452f3cfc68b7', null, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Create', 'employees.create'),
('e4e68398-da67-438f-9fb0-07a779b504a0', null, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Edit', 'employees.edit'),
('33279d12-6daa-41d3-a037-9f805e8ebf61', null, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Delete', 'employees.delete');

INSERT INTO public.role_permissions (id, deleted_at, created_at, updated_at, role_id, permission_id) VALUES
('f73306cc-5b01-49bb-87af-a2c2af14bfae', null, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '01ed1076-4028-4fdf-9a92-cb57a8e041af'),
('ef3e5ee2-69b5-44a6-85fd-0b8e5329ae6b', null, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'a7549f30-987c-47d1-9266-452f3cfc68b7'),
('6ad4a140-86ee-442b-8a7a-a0566dbab1f4', null, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'e4e68398-da67-438f-9fb0-07a779b504a0'),
('531e689e-9be2-442b-ac31-111f12e11a89', null, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '33279d12-6daa-41d3-a037-9f805e8ebf61'),
('75012959-344d-449f-a7ec-00023d68b32b', null, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', 'd796884d-a8c4-4525-81e7-54a3b6099eac', '01ed1076-4028-4fdf-9a92-cb57a8e041af');

INSERT INTO public.positions (id, deleted_at, created_at, updated_at, name, code) VALUES 
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Frontend', 'frontend'),
('d796884d-a8c4-4525-81e7-54a3b6099eac', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Backend', 'backend'),
('dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Devops', 'devops'),
('01fb6322-d727-47e3-a242-5039ea4732fc', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Blockchain', 'blockchain'),
('39735742-829b-47f3-8f9d-daf0983914e5', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Project-Management', 'project-management');

INSERT INTO "public"."countries" ("id", "deleted_at", "created_at", "updated_at", "name", "code", "cities") VALUES 
('4ef64490-c906-4192-a7f9-d2221dadfe4c',NULL,'2022-11-08 08:06:56.068148','2022-11-08 08:06:56.068148','Vietnam','+84','["Hồ Chí Minh", "An Giang", "Bà Rịa-Vũng Tàu", "Bình Dương", "Bình Định", "Bình Phước", "Bình Thuận", "Bạc Liêu", "Bắc Giang", "Bắc Kạn", "Bắc Ninh", "Bến Tre", "Cao Bằng", "Cà Mau", "Cần Thơ", "Điện Biên", "Đà Nẵng", "Đắk Lắk", "Đồng Nai", "Đắk Nông", "Đồng Tháp", "Gia Lai", "Hoà Bình", "Hà Giang", "Hà Nam", "Hà Nội", "Hà Tĩnh", "Hải Dương", "Hải Phòng", "Hậu Giang", "Hưng Yên", "Khánh Hòa", "Kiên Giang", "Kon Tum", "Lai Châu", "Lâm Đồng", "Lạng Sơn", "Lào Cai", "Long An", "Nam Định", "Nghệ An", "Ninh Bình", "Ninh Thuận", "Phú Thọ", "Phú Yên", "Quảng Bình", "Quảng Nam", "Quảng Ngãi", "Quảng Ninh", "Quảng Trị", "Sóc Trăng", "Sơn La", "Thanh Hóa", "Thái Bình", "Thái Nguyên", "Thừa Thiên Huế", "Tiền Giang", "Trà Vinh", "Tuyên Quang", "Tây Ninh", "Vĩnh Long", "Vĩnh Phúc", "Yên Bái"]'),
('da9031ce-0d6e-4344-b97a-a2c44c66153e',NULL,'2022-11-08 08:08:09.881727','2022-11-08 08:08:09.881727','Singapore','+65','["Singapore"]');

INSERT INTO public.seniorities (id, deleted_at, created_at, updated_at, name, code) VALUES
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Fresher', 'fresher'),
('d796884d-a8c4-4525-81e7-54a3b6099eac', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Junior', 'junior'),
('dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Mid', 'mid'),
('01fb6322-d727-47e3-a242-5039ea4732fc', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Senior', 'senior'),
('01fb6322-d727-47e3-a242-5039ea4732fd', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Staff', 'staff'),
('39735742-829b-47f3-8f9d-daf0983914e5', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Principal', 'principal');

INSERT INTO public.chapters (id, deleted_at, created_at, updated_at, name, code) VALUES
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Web', 'web'),
('d796884d-a8c4-4525-81e7-54a3b6099eac', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Backend', 'backend'),
('dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'BlockChain', 'blockchain'),
('01fb6322-d727-47e3-a242-5039ea4732fc', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'DevOps', 'devops'),
('01fb6322-d727-47e3-a242-5039ea4732fd', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Mobile', 'mobile'),
('01fb6322-d727-47e3-a242-5039ea4732fe', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'QA', 'qa'),
('39735742-829b-47f3-8f9d-daf0983914e5', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'PM', 'pm');

INSERT INTO public.stacks (id, deleted_at, created_at, updated_at, name, code, avatar) VALUES
('0ecf47c8-cca4-4c30-94bb-054b1124c44f', null, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'Golang', 'golang', null),
('fa0f4e46-7eab-4e5c-9d31-30489e69fe2e', null, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'React', 'react', null),
('b403ef95-4269-4830-bbb6-8e56e5ec0af4', null, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'Google Cloud', 'gcloud', null);

INSERT INTO public.projects (id, deleted_at, created_at, updated_at, name, type, start_date, end_date, status) VALUES
('8dc3be2e-19a4-4942-8a79-56db391a0b15', null, '2022-11-11 18:06:56.362902', '2022-11-11 18:06:56.362902', 'Fortress', 'dwarves', '2022-11-01', null, 'active'),
('dfa182fc-1d2d-49f6-a877-c01da9ce4207', null, '2022-11-11 18:06:56.362902', '2022-11-11 18:06:56.362902', 'Lorem ipsum', 'time-material', '2022-07-06', null, 'active');

INSERT INTO public.project_slots (id, deleted_at, created_at, updated_at, project_id, seniority_id, upsell_person_id, position, deployment_type, rate, discount, status) VALUES
('f32d08ca-8863-4ab3-8c84-a11849451eb7', null, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', null,'frontend', 'official', 5000, 0, 'active'),
('bdc64b18-4c5f-4025-827a-f5b91d599dc7', null, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', null,'backend', 'shadow', 4000, 0, 'active'),
('1406fcce-6f90-4e0f-bea1-c373e2b2b5b1', null, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', null,'backend', 'official', 3000, 3, 'active'),
('b25bd3fa-eb6d-49d5-b278-7aacf4594f79', null, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', null,'frontend', 'official', 3000, 3, 'active'),
('ce379dc0-95be-471a-9227-8e045a5630af', null, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', null,'project-management', 'shadow', 4000, 0, 'active');

INSERT INTO public.project_members (id, deleted_at, created_at, updated_at, project_id, project_slot_id ,employee_id, seniority_id, joined_date, left_date, position, rate, discount, status,deployment_type, upsell_person_id) VALUES
('cb889a9c-b20c-47ee-83b8-44b6d1721aca', null, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'f32d08ca-8863-4ab3-8c84-a11849451eb7', '2655832e-f009-4b73-a535-64c3a22e558f', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', null, 'frontend', 5000, 0, 'active', 'official', '608ea227-45a5-4c8a-af43-6c7280d96340'),
('7310b51a-3855-498b-99ab-41aa82934269', null, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'bdc64b18-4c5f-4025-827a-f5b91d599dc7', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', null, 'backend', 4000, 0, 'active', 'shadow', '608ea227-45a5-4c8a-af43-6c7280d96340'),
('35149aab-0506-4eb3-9300-c706ccbf2bde', null, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '1406fcce-6f90-4e0f-bea1-c373e2b2b5b1', '8d7c99c0-3253-4286-93a9-e7554cb327ef', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', null, 'backend', 3000, 3, 'active', 'official', '608ea227-45a5-4c8a-af43-6c7280d96340'),
('5a9a07aa-e8f3-4b62-b9ad-0f057866dc6c', null, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'b25bd3fa-eb6d-49d5-b278-7aacf4594f79', 'eeae589a-94e3-49ac-a94c-fcfb084152b2', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', null, 'frontend', 3000, 3, 'active', 'official', '608ea227-45a5-4c8a-af43-6c7280d96340'),
('fcd5c16f-40fd-48b6-9649-410691373eea', null, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'ce379dc0-95be-471a-9227-8e045a5630af', '608ea227-45a5-4c8a-af43-6c7280d96340', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', null, 'project-management', 4000, 0, 'active', 'shadow', '608ea227-45a5-4c8a-af43-6c7280d96340');

INSERT INTO public.project_heads (id, deleted_at, created_at, updated_at, project_id, employee_id, joined_date, left_date, commission_rate, position) VALUES
('528433a5-4001-408d-bd0c-3b032eef0a70', null, '2022-11-11 18:27:03.906845', '2022-11-11 18:27:03.906845', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '2655832e-f009-4b73-a535-64c3a22e558f', '2022-11-01', null, 1, 'account-manager'),
('14d1d9f0-5f0f-49d5-8309-ac0a40de013d', null, '2022-11-11 18:27:03.906845', '2022-11-11 18:27:03.906845', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '8d7c99c0-3253-4286-93a9-e7554cb327ef', '2022-11-01', null, 1, 'technical-lead'),
('e17b6bfc-79b0-4d65-ac88-559d8c597d2e', null, '2022-11-11 18:27:03.906845', '2022-11-11 18:27:03.906845', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '608ea227-45a5-4c8a-af43-6c7280d96340', '2022-11-02', null, 2, 'sale-person'),
('72578fb2-921b-4c8b-b998-a4aec73da809', null, '2022-11-11 18:27:03.906845', '2022-11-11 18:27:03.906845', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', '2022-11-01', null, 2.5, 'technical-lead'),
('eca9fed2-21cf-4c75-9ab5-f6e5a13a0a75', null, '2022-11-11 18:27:03.906845', '2022-11-11 18:27:03.906845', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '2655832e-f009-4b73-a535-64c3a22e558f', '2022-11-01', null, 0.5, 'delivery-manager');

INSERT INTO public.project_stacks (id, deleted_at, created_at, updated_at, project_id, stack_id) VALUES
('6b89cc06-33ca-4f48-825a-ab20da5cb287', null, '2022-11-11 18:39:34.923619', '2022-11-11 18:39:34.923619', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('da982888-1727-4c06-8e8a-86aff1f80f89', null, '2022-11-11 18:39:34.923619', '2022-11-11 18:39:34.923619', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'fa0f4e46-7eab-4e5c-9d31-30489e69fe2e'),
('371d797f-a639-4869-8026-67e682c6da4a', null, '2022-11-11 18:39:34.923619', '2022-11-11 18:39:34.923619', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'b403ef95-4269-4830-bbb6-8e56e5ec0af4');

INSERT INTO public.employee_stacks (id,deleted_at,created_at,updated_at,employee_id,stack_id) VALUES
('a9921fe8-3b77-46ad-a76d-172cb85e10a5',NULL,'2022-11-16 11:28:01.048359','2022-11-16 11:28:01.048359','ecea9d15-05ba-4a4e-9787-54210e3b98ce','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('94de566e-9d1c-4ce7-a16a-0deaf3a45f9a',NULL,'2022-11-16 11:28:01.051869','2022-11-16 11:28:01.051869','ecea9d15-05ba-4a4e-9787-54210e3b98ce','fa0f4e46-7eab-4e5c-9d31-30489e69fe2e'),
('8eb2ffa5-68aa-4b0c-b113-e8b37297031e',NULL,'2022-11-16 11:28:01.053014','2022-11-16 11:28:01.053014','2655832e-f009-4b73-a535-64c3a22e558f','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('e2c3d67b-47e3-4020-8684-2971b3e12b9b',NULL,'2022-11-16 11:28:01.054087','2022-11-16 11:28:01.054087','2655832e-f009-4b73-a535-64c3a22e558f','b403ef95-4269-4830-bbb6-8e56e5ec0af4'),
('c3fd413f-1a7e-4be9-939b-97745ae98715',NULL,'2022-11-16 11:28:01.059009','2022-11-16 11:28:01.059009','8d7c99c0-3253-4286-93a9-e7554cb327ef','fa0f4e46-7eab-4e5c-9d31-30489e69fe2e'),
('da9f4cb2-0230-4c28-babb-7ed7ca9d5682',NULL,'2022-11-16 11:28:01.060256','2022-11-16 11:28:01.060256','8d7c99c0-3253-4286-93a9-e7554cb327ef','b403ef95-4269-4830-bbb6-8e56e5ec0af4'),
('4c137480-98af-4dcb-a7c0-ec56cc02468b',NULL,'2022-11-16 11:28:01.06115','2022-11-16 11:28:01.06115','eeae589a-94e3-49ac-a94c-fcfb084152b2','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('37ba3955-2917-4f09-a647-74486e56b28f',NULL,'2022-11-16 11:28:01.061939','2022-11-16 11:28:01.061939','608ea227-45a5-4c8a-af43-6c7280d96340','b403ef95-4269-4830-bbb6-8e56e5ec0af4');

INSERT INTO public.employee_positions (id, deleted_at, created_at, updated_at, employee_id, position_id) VALUES
('e61bf924-1eff-426f-8f71-ab49446134fb',NULL,'2022-11-16 02:45:17.982659','2022-11-16 02:45:17.982659','2655832e-f009-4b73-a535-64c3a22e558f','11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('d7b49b9e-4b51-4e3a-a7be-b94143215f7a',NULL,'2022-11-16 02:45:17.995921','2022-11-16 02:45:17.995921','2655832e-f009-4b73-a535-64c3a22e558f','d796884d-a8c4-4525-81e7-54a3b6099eac'),
('4bb1ea40-c002-4fd9-8f60-82b3aa9459d7',NULL,'2022-11-16 02:45:17.997153','2022-11-16 02:45:17.997153','ecea9d15-05ba-4a4e-9787-54210e3b98ce','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('8fb91cd7-f461-49fb-8b7b-9ef9b524789d',NULL,'2022-11-16 02:45:17.998377','2022-11-16 02:45:17.998377','ecea9d15-05ba-4a4e-9787-54210e3b98ce','01fb6322-d727-47e3-a242-5039ea4732fc');
