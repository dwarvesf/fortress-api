INSERT INTO public.countries (id, deleted_at, created_at, updated_at, name, code, cities) VALUES
('4ef64490-c906-4192-a7f9-d2221dadfe4c',NULL,'2022-11-08 08:06:56.068148','2022-11-08 08:06:56.068148','Vietnam','+84','["Hồ Chí Minh", "An Giang", "Bà Rịa-Vũng Tàu", "Bình Dương", "Bình Định", "Bình Phước", "Bình Thuận", "Bạc Liêu", "Bắc Giang", "Bắc Kạn", "Bắc Ninh", "Bến Tre", "Cao Bằng", "Cà Mau", "Cần Thơ", "Điện Biên", "Đà Nẵng", "Đắk Lắk", "Đồng Nai", "Đắk Nông", "Đồng Tháp", "Gia Lai", "Hoà Bình", "Hà Giang", "Hà Nam", "Hà Nội", "Hà Tĩnh", "Hải Dương", "Hải Phòng", "Hậu Giang", "Hưng Yên", "Khánh Hòa", "Kiên Giang", "Kon Tum", "Lai Châu", "Lâm Đồng", "Lạng Sơn", "Lào Cai", "Long An", "Nam Định", "Nghệ An", "Ninh Bình", "Ninh Thuận", "Phú Thọ", "Phú Yên", "Quảng Bình", "Quảng Nam", "Quảng Ngãi", "Quảng Ninh", "Quảng Trị", "Sóc Trăng", "Sơn La", "Thanh Hóa", "Thái Bình", "Thái Nguyên", "Thừa Thiên Huế", "Tiền Giang", "Trà Vinh", "Tuyên Quang", "Tây Ninh", "Vĩnh Long", "Vĩnh Phúc", "Yên Bái"]'),
('da9031ce-0d6e-4344-b97a-a2c44c66153e',NULL,'2022-11-08 08:08:09.881727','2022-11-08 08:08:09.881727','Singapore','+65','["Singapore"]');

INSERT INTO public.currencies (id, deleted_at, created_at, updated_at, name, symbol, locale, type) VALUES
('06a699ed-618b-400b-ac8c-8739956fa8e7', NULL, '2019-02-20 04:24:14.967209+00', '2019-02-20 04:24:14.967209+00', 'GBP', '£', 'en-gb', 'fiat'),
('0a6f4a2e-a097-4f7e-ae65-bfee3298e5cc', NULL, '2020-11-23 08:59:49.802559+00', '2020-11-23 08:59:49.802559+00', 'TEN', NULL, NULL, 'crypto'),
('1c7dcbe2-6984-461d-8ed9-537676f2b590', NULL, '2021-07-21 18:10:24.194492+00', '2021-07-21 18:10:24.194492+00', 'USD', '$', 'en-us', 'crypto'),
('283e559f-a7e7-4aa7-9e08-4806d6c07016', NULL, '2019-02-20 04:24:14.967209+00', '2019-02-20 04:24:14.967209+00', 'EUR', '€', 'en-gb', 'fiat'),
('2de81125-a947-4fea-a006-2c60e7ec01ed', NULL, '2020-02-28 07:18:02.349700+00', '2020-02-28 07:18:02.349700+00', 'CAD', 'c$', 'en-ca', 'fiat'),
('7037bdb6-584e-4e35-996d-ef28a243f48a', NULL, '2019-02-11 06:14:51.305496+00', '2019-02-11 06:14:51.305496+00', 'VND', 'đ', 'vi-vn', 'fiat'),
('bf256e69-28b0-4d9f-bf48-3662854157a9', NULL, '2019-10-28 14:14:52.302051+00', '2019-10-28 14:14:52.302051+00', 'SGD', 's$', 'en-sg', 'fiat'),
('f00498e4-7a4c-4f61-b126-b84b5faeee06', NULL, '2019-02-11 06:14:51.305496+00', '2019-02-11 06:14:51.305496+00', 'USD', '$', 'en-us', 'fiat');

INSERT INTO public.positions (id, deleted_at, created_at, updated_at, name, code) VALUES
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Frontend', 'frontend'),
('d796884d-a8c4-4525-81e7-54a3b6099eac', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Backend', 'backend'),
('dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Devops', 'devops'),
('01fb6322-d727-47e3-a242-5039ea4732fc', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Blockchain', 'blockchain'),
('39735742-829b-47f3-8f9d-daf0983914e5', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Project Manager', 'project-manager'),
('65c67b84-fbec-406f-bcae-a947a6f7f12a', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Data Engineer', 'data-engineer'),
('0590d7f2-22ba-4f53-bc51-16e5aa775beb', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Data Analyst', 'data-analyst'),
('ce2b8c4d-2ab6-4a32-82c2-e114d428fb1e', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Quality Assurance', 'quality-assurance'),
('86d9d3df-a329-4013-b85a-452e4c9a3182', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Product Owner', 'product-owner'),
('0b3fb4c7-75e8-4a2f-a20c-2423c7c80131', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Scrum Master', 'scrum-master'),
('f5d0163a-f94b-4118-a49f-04076e7a0f8b', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Business Analyst', 'business-analyst'),
('7214530a-1b41-4dc6-ad02-4b6588e403df', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Mobile', 'mobile'),
('d92cef88-4277-4542-985f-5aa9ed5f39b9', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'UX Design', 'ux-design'),
('002abc17-de77-4636-b80d-6f6acd3de679', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'UI Design', 'ui-design'),
('742f8d40-6078-4173-bc8a-fecf2a463feb', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Product-Design', 'product-design'),
('095248c4-b271-4186-95ef-f398a6a6e430', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Lead', 'lead'),
('a786c042-ac58-447d-ba43-eee887526367', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Engineering Lead', 'engineering-lead'),
('396f07b1-419f-474a-bbf4-b91c11666137', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Engineering Manager', 'engineering-manager'),
('100de34f-df16-4f60-9951-d3f581022cff', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Head Of Engineering', 'head-of-engineering'),
('0d8ea8e0-392a-46d0-a53c-4d84d6f47366', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Product Manager', 'product-manager'),
('309a3a1e-8eb5-4777-849e-d4d9924797a0', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Fullstack', 'fullstack');

INSERT INTO public.seniorities (id, deleted_at, created_at, updated_at, name, code) VALUES
('d796884d-a8c4-4525-81e7-54a3b6099eac', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Junior', 'junior'),
('dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Mid', 'mid'),
('01fb6322-d727-47e3-a242-5039ea4732fc', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Senior', 'senior'),
('01fb6322-d727-47e3-a242-5039ea4732fd', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Staff', 'staff'),
('39735742-829b-47f3-8f9d-daf0983914e5', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Principal', 'principal');

INSERT INTO public.stacks (id, deleted_at, created_at, updated_at, name, code, avatar) VALUES
('0ecf47c8-cca4-4c30-94bb-054b1124c44f', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'Golang', 'golang', NULL),
('fa0f4e46-7eab-4e5c-9d31-30489e69fe2e', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'React', 'react', NULL),
('b403ef95-4269-4830-bbb6-8e56e5ec0af4', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'Elixir', 'elixir', NULL),
('4e6d4c88-7e2e-478e-811c-17b32b32ec94', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'AWS', 'aws', NULL),
('426fa044-465e-453c-8c0d-b1d34843e108', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'GCP', 'gcp', NULL),
('44bb9de8-d0dc-4126-a8ab-36247108ab95', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'Java', 'java', NULL),
('00cd96a6-a7f3-402e-af94-256c1f23ab75', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'TypeScript', 'typescript', NULL),
('588b36ea-3695-4384-83d7-77462474ebcf', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'Swift', 'swift', NULL),
('eef98cf6-9d58-4d00-aebc-2b8a4bb439a1', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'Android', 'android', NULL),
('54a3ac80-3ba0-4aeb-8852-904e2a1d7263', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'Flutter', 'flutter', NULL),
('d7a4f73e-35f4-425b-b380-6fa19199fc3e', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'React Native', 'react-native', NULL),
('569e50b4-0aa9-4406-aa99-fc9472ce5723', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'Angular', 'angular', NULL),
('8944def6-6455-4e90-bbc9-3ab7a61f2ab3', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'Vue', 'vue', NULL);

INSERT INTO public.questions (id, deleted_at, created_at, updated_at, category, subcategory, content, type, "order", domain) VALUES 
('da5dbdd5-8e1e-4ae7-8bb8-ab007f2580aa', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'Does this employee effectively communicate with others?', 'general', 1, NULL),
('7d95e035-81d6-49d7-bed4-3a83bf2e34d6', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'How effective of a leader is this person?', 'general', 2, NULL),
('d36e84c5-d7a4-4d5f-ada1-f6b9ddb58f51', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'Does this person find creative solutions?', 'general', 3, NULL),
('f03432ba-c024-467e-8059-a5bb2b7f783d', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'How would you rate the quality of the employee''s work?', 'general', 4, NULL),
('d2bb48c1-e8d6-4946-a372-8499907b7328', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'How well does this person set and meet deadlines?', 'general', 5, NULL),
('be86ce52-803b-403f-b059-1a69492fe3d4', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'How well does this person embody our culture?', 'general', 6, NULL),
('51eab8c7-61ba-4c56-be39-b72eb6b89a52', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'If you could give this person one piece of constructive advice to make them more effective in their role?', 'general', 7, NULL),
('4e71821e-e8d7-4fc9-9f02-7b4de4efa7f8', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'engagement', 'I know what is expected of me at work.', 'likert-scale', 1, 'engagement'),
('ba00c397-f8ad-4cbd-ba63-02ac903d0886', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'I have the materials and equipment I need to do my work right.', 'likert-scale', 2, 'engagement'),
('8c25c7c1-5148-4b85-92b2-5db919b0a118', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'At work, I have the opportunity to do what I do best every day.', 'likert-scale', 3, 'engagement'),
('6c6bf3d6-46cd-46d3-9ea6-f382a5052588', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'In the last seven days, I have received recognition or praise for doing good work.', 'likert-scale', 4, 'engagement'),
('c7f2081a-11e6-4a2e-9e0c-852faaaf3d16', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'My supervisor, or someone at work, seems to care about me as a person.', 'likert-scale', 5, 'engagement'),
('729cf731-ba4c-4b62-8aa6-8f4a9f4507b8', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'There is someone at work who encourages my development.', 'likert-scale', 6, 'engagement'),
('ee16d005-4f9f-4a0c-93b0-43db5f8edbb3', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'At work, my opinions seem to count.', 'likert-scale', 7, 'engagement'),
('989d8e60-4b71-46d6-8d56-77ca8d37877c', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'The mission or purpose of my company makes me feel my job is important.', 'likert-scale', 8, 'engagement'),
('41db6b66-6942-43b3-b1fc-568ba4579762', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'My associates or fellow employees are committed to doing quality work.', 'likert-scale', 9, 'engagement'),
('1b9a91c7-7e67-4693-bb43-74170aaa0e3f', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'I have a best friend at work.', 'likert-scale', 10, 'engagement'),
('edb5f946-1d2b-4029-9b1f-57483e92be60', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'In the last six months, someone at work has talked to me about my progress.', 'likert-scale', 11, 'engagement'),
('8d5f0723-e867-430e-82fc-26f8809140b2', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'This last year, I have had opportunities at work to learn and grow.', 'likert-scale', 12, 'engagement'),
('e703b6ee-e71a-4897-8716-f3811963ffab', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'work', 'How do you feel about your workload?', 'likert-scale', 1, 'workload'),
('d9bf74dd-6f25-44e3-b83c-9b7fb19af548', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'work', 'Do you think the team can make the deadline?', 'likert-scale', 2, 'deadline'),
('627b2be7-2dcb-4574-b8db-24d67f06e7b0', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'work', 'How much did you learn from your work and your team this week?', 'likert-scale', 3, 'learning');
