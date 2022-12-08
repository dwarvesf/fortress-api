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
('39735742-829b-47f3-8f9d-daf0983914e5', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Project-Management', 'project-management');

INSERT INTO public.seniorities (id, deleted_at, created_at, updated_at, name, code) VALUES
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Fresher', 'fresher'),
('d796884d-a8c4-4525-81e7-54a3b6099eac', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Junior', 'junior'),
('dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Mid', 'mid'),
('01fb6322-d727-47e3-a242-5039ea4732fc', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Senior', 'senior'),
('01fb6322-d727-47e3-a242-5039ea4732fd', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Staff', 'staff'),
('39735742-829b-47f3-8f9d-daf0983914e5', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Principal', 'principal');

INSERT INTO public.stacks (id, deleted_at, created_at, updated_at, name, code, avatar) VALUES
('0ecf47c8-cca4-4c30-94bb-054b1124c44f', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'Golang', 'golang', NULL),
('fa0f4e46-7eab-4e5c-9d31-30489e69fe2e', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'React', 'react', NULL),
('b403ef95-4269-4830-bbb6-8e56e5ec0af4', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'Google Cloud', 'gcloud', NULL);

INSERT INTO public.questions (id, deleted_at, created_at, updated_at, category, subcategory, content, type, "order") VALUES 
('da5dbdd5-8e1e-4ae7-8bb8-ab007f2580aa', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'Does this employee effectively communicate with others?', 'general', 1),
('7d95e035-81d6-49d7-bed4-3a83bf2e34d6', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'How effective of a leader is this person', 'general', 2),
('d36e84c5-d7a4-4d5f-ada1-f6b9ddb58f51', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'Does this person find creative solutions', 'general', 3),
('f03432ba-c024-467e-8059-a5bb2b7f783d', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'How would you rate the quality of the employee''s work?', 'general', 4),
('d2bb48c1-e8d6-4946-a372-8499907b7328', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'How well does this person set and meet deadlines?', 'general', 5),
('be86ce52-803b-403f-b059-1a69492fe3d4', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'How well does this person embody our culture?', 'general', 6),
('51eab8c7-61ba-4c56-be39-b72eb6b89a52', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'If you could give this person one piece of constructive advice to make them more effective in their role', 'general', 7),
('4e71821e-e8d7-4fc9-9f02-7b4de4efa7f8', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'engagement', 'I know what is expected of me at work.', 'likert-scale', 1),
('ba00c397-f8ad-4cbd-ba63-02ac903d0886', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'I have the materials and equipment I need to do my work right.', 'likert-scale', 2),
('8c25c7c1-5148-4b85-92b2-5db919b0a118', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'At work, I have the opportunity to do what I do best every day.', 'likert-scale', 3),
('6c6bf3d6-46cd-46d3-9ea6-f382a5052588', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'In the last seven days, I have received recognition or praise for doing good work.', 'likert-scale', 4),
('c7f2081a-11e6-4a2e-9e0c-852faaaf3d16', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'My supervisor, or someone at work, seems to care about me as a person.', 'likert-scale', 5),
('729cf731-ba4c-4b62-8aa6-8f4a9f4507b8', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'There is someone at work who encourages my development.', 'likert-scale', 6),
('ee16d005-4f9f-4a0c-93b0-43db5f8edbb3', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'At work, my opinions seem to count.', 'likert-scale', 7),
('989d8e60-4b71-46d6-8d56-77ca8d37877c', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'The mission or purpose of my company makes me feel my job is important.', 'likert-scale', 8),
('41db6b66-6942-43b3-b1fc-568ba4579762', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'My associates or fellow employees are committed to doing quality work.', 'likert-scale', 9),
('1b9a91c7-7e67-4693-bb43-74170aaa0e3f', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'I have a best friend at work.', 'likert-scale', 10),
('edb5f946-1d2b-4029-9b1f-57483e92be60', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'In the last six months, someone at work has talked to me about my progress.', 'likert-scale', 11),
('8d5f0723-e867-430e-82fc-26f8809140b2', NULL, '2022-12-06 03:07:13.021661', '2022-12-06 03:07:13.021661', 'survey', 'engagement', 'This last year, I have had opportunities at work to learn and grow.', 'likert-scale', 12);
