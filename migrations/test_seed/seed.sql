INSERT INTO public.roles (id, deleted_at, created_at, updated_at, name, code) VALUES
    ('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Admin', 'admin'),
    ('d796884d-a8c4-4525-81e7-54a3b6099eac', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Member', 'member');

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

INSERT INTO public.projects (id, deleted_at, created_at, updated_at, name, type, start_date, end_date, status, country_id, client_email, project_email) VALUES
('8dc3be2e-19a4-4942-8a79-56db391a0b15', null, '2022-11-11 18:06:56.362902', '2022-11-11 18:06:56.362902', 'Fortress', 'dwarves', '2022-11-01', null, 'active', '4ef64490-c906-4192-a7f9-d2221dadfe4c', 'fortress@gmai.com', 'fortress@d.foundation'),
('dfa182fc-1d2d-49f6-a877-c01da9ce4207', null, '2022-11-11 18:06:56.362902', '2022-11-11 18:06:56.362902', 'Lorem ipsum', 'time-material', '2022-07-06', null, 'active', '4ef64490-c906-4192-a7f9-d2221dadfe4c', 'project@gmai.com', 'project@d.foundation');

INSERT INTO public.stacks (id, deleted_at, created_at, updated_at, name, code, avatar) VALUES
('0ecf47c8-cca4-4c30-94bb-054b1124c44f', null, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'Golang', 'golang', null),
('fa0f4e46-7eab-4e5c-9d31-30489e69fe2e', null, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'React', 'react', null),
('b403ef95-4269-4830-bbb6-8e56e5ec0af4', null, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'Google Cloud', 'gcloud', null);

INSERT INTO public.employees (id, deleted_at, created_at, updated_at, full_name, display_name, gender, team_email, personal_email, avatar, phone_number, address, mbti, horoscope, date_of_birth, working_status, passport_photo_front, passport_photo_back, identity_card_photo_front, identity_card_photo_back, joined_date, left_date, basecamp_id, basecamp_attachable_sgid, gitlab_id, github_id, discord_id, wise_recipient_email, wise_recipient_name, wise_recipient_id, wise_account_number, wise_currency, local_bank_branch, local_bank_number, local_bank_currency, local_branch_name, local_bank_recipient_name,seniority_id, chapter_id, line_manager_id) VALUES
('2655832e-f009-4b73-a535-64c3a22e558f', NULL, '2022-11-02 09:52:34.586566', '2022-11-02 09:52:34.586566', 'Phạm Đức Thành', 'Thanh Pham', 'Male', 'thanh@d.foundation', 'thanhpham123@gmail.com', 'https://s3-ap-southeast-1.amazonaws.com/fortress-images/5153574695663955944.png', '0123456788', 'Phan Huy Ich, Tan Binh District, Ho Chi Minh, Vietnam', 'INTJ-A', 'Libra', '1990-01-02', 'contractor', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '39735742-829b-47f3-8f9d-daf0983914e5', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce'),
('ecea9d15-05ba-4a4e-9787-54210e3b98ce', NULL, '2022-11-02 09:50:55.320669', '2022-11-02 09:50:55.320669', 'Nguyễn Hoàng Huy', 'Huy Nguyen', 'Male', 'huy123@d.foundation', 'hoanghuy123@gmail.com', 'https://s3-ap-southeast-1.amazonaws.com/fortress-images/2830497479497502617.png', '0123456789', 'chung cư Sunview Town, Gò Dưa, Thủ Đức', 'Defender', 'Virgo', '1990-01-02', 'probation', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '39735742-829b-47f3-8f9d-daf0983914e5', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce');

INSERT INTO public.employee_positions (id, deleted_at, created_at, updated_at, employee_id, position_id) VALUES
('e61bf924-1eff-426f-8f71-ab49446134fb',NULL,'2022-11-16 02:45:17.982659','2022-11-16 02:45:17.982659','2655832e-f009-4b73-a535-64c3a22e558f','11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('d7b49b9e-4b51-4e3a-a7be-b94143215f7a',NULL,'2022-11-16 02:45:17.995921','2022-11-16 02:45:17.995921','2655832e-f009-4b73-a535-64c3a22e558f','d796884d-a8c4-4525-81e7-54a3b6099eac'),
('4bb1ea40-c002-4fd9-8f60-82b3aa9459d7',NULL,'2022-11-16 02:45:17.997153','2022-11-16 02:45:17.997153','ecea9d15-05ba-4a4e-9787-54210e3b98ce','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('8fb91cd7-f461-49fb-8b7b-9ef9b524789d',NULL,'2022-11-16 02:45:17.998377','2022-11-16 02:45:17.998377','ecea9d15-05ba-4a4e-9787-54210e3b98ce','01fb6322-d727-47e3-a242-5039ea4732fc');

INSERT INTO public.employee_stacks (id,deleted_at,created_at,updated_at,employee_id,stack_id) VALUES
('a9921fe8-3b77-46ad-a76d-172cb85e10a5',NULL,'2022-11-16 11:28:01.048359','2022-11-16 11:28:01.048359','ecea9d15-05ba-4a4e-9787-54210e3b98ce','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('94de566e-9d1c-4ce7-a16a-0deaf3a45f9a',NULL,'2022-11-16 11:28:01.051869','2022-11-16 11:28:01.051869','ecea9d15-05ba-4a4e-9787-54210e3b98ce','fa0f4e46-7eab-4e5c-9d31-30489e69fe2e'),
('8eb2ffa5-68aa-4b0c-b113-e8b37297031e',NULL,'2022-11-16 11:28:01.053014','2022-11-16 11:28:01.053014','2655832e-f009-4b73-a535-64c3a22e558f','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('e2c3d67b-47e3-4020-8684-2971b3e12b9b',NULL,'2022-11-16 11:28:01.054087','2022-11-16 11:28:01.054087','2655832e-f009-4b73-a535-64c3a22e558f','b403ef95-4269-4830-bbb6-8e56e5ec0af4');


INSERT INTO public.project_slots (id, deleted_at, created_at, updated_at, project_id, seniority_id, upsell_person_id, position, deployment_type, rate, discount, status) VALUES
('f32d08ca-8863-4ab3-8c84-a11849451eb7', null, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', null,'frontend', 'official', 5000, 0, 'active'),
('bdc64b18-4c5f-4025-827a-f5b91d599dc7', null, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', null,'backend', 'shadow', 4000, 0, 'active');

INSERT INTO public.project_members (id, deleted_at, created_at, updated_at, project_id, project_slot_id ,employee_id, seniority_id, joined_date, left_date, rate, discount, status,deployment_type, upsell_person_id) VALUES
('cb889a9c-b20c-47ee-83b8-44b6d1721aca', null, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'f32d08ca-8863-4ab3-8c84-a11849451eb7', '2655832e-f009-4b73-a535-64c3a22e558f', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', null, 5000, 0, 'active', 'official', '2655832e-f009-4b73-a535-64c3a22e558f'),
('7310b51a-3855-498b-99ab-41aa82934269', null, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'bdc64b18-4c5f-4025-827a-f5b91d599dc7', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', null, 4000, 0, 'active', 'shadow', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce');

INSERT INTO public.project_stacks (id, project_id, stack_id) VALUES
('e7a0dc5e-7da6-48df-879a-f392002732c3', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '0ecf47c8-cca4-4c30-94bb-054b1124c44f');

INSERT INTO public.currencies (id, deleted_at, created_at, updated_at, name, symbol, locale, type) VALUES
('06a699ed-618b-400b-ac8c-8739956fa8e7', NULL, '2019-02-20 04:24:14.967209+00', '2019-02-20 04:24:14.967209+00', 'GBP', '£', 'en-gb', 'fiat'),
('0a6f4a2e-a097-4f7e-ae65-bfee3298e5cc', NULL, '2020-11-23 08:59:49.802559+00', '2020-11-23 08:59:49.802559+00', 'TEN', NULL, NULL, 'crypto'),
('1c7dcbe2-6984-461d-8ed9-537676f2b590', NULL, '2021-07-21 18:10:24.194492+00', '2021-07-21 18:10:24.194492+00', 'USD', '$', 'en-us', 'crypto'),
('283e559f-a7e7-4aa7-9e08-4806d6c07016', NULL, '2019-02-20 04:24:14.967209+00', '2019-02-20 04:24:14.967209+00', 'EUR', '€', 'en-gb', 'fiat'),
('2de81125-a947-4fea-a006-2c60e7ec01ed', NULL, '2020-02-28 07:18:02.349700+00', '2020-02-28 07:18:02.349700+00', 'CAD', 'c$', 'en-ca', 'fiat'),
('7037bdb6-584e-4e35-996d-ef28a243f48a', NULL, '2019-02-11 06:14:51.305496+00', '2019-02-11 06:14:51.305496+00', 'VND', 'đ', 'vi-vn', 'fiat'),
('bf256e69-28b0-4d9f-bf48-3662854157a9', NULL, '2019-10-28 14:14:52.302051+00', '2019-10-28 14:14:52.302051+00', 'SGD', 's$', 'en-sg', 'fiat'),
('f00498e4-7a4c-4f61-b126-b84b5faeee06', NULL, '2019-02-11 06:14:51.305496+00', '2019-02-11 06:14:51.305496+00', 'USD', '$', 'en-us', 'fiat');
