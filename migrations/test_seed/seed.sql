INSERT INTO public.roles (id, deleted_at, created_at, updated_at, name, code, level) VALUES
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Admin', 'admin', 1),
('d796884d-a8c4-4525-81e7-54a3b6099eac', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Member', 'member', 2);

INSERT INTO public.permissions (id, deleted_at, created_at, updated_at, name, code) VALUES
('01ed1076-4028-4fdf-9a92-cb57a8e041af', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read', 'employees.read'),
('a7549f30-987c-47d1-9266-452f3cfc68b7', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Create', 'employees.create'),
('e4e68398-da67-438f-9fb0-07a779b504a0', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Edit', 'employees.edit'),
('33279d12-6daa-41d3-a037-9f805e8ebf61', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Delete', 'employees.delete'),
('09a5fa4c-ad07-4dec-a16d-34e0f567ef1d', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Create', 'projects.create'),
('35bb795a-0b9f-428c-861d-48c1e8d4e73a', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Read', 'projects.read'),
('80a2c800-02ea-4264-9289-57b92e911097', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Edit', 'projects.edit'),
('6797aeac-7d5e-4216-b260-a8a3f62d3cf6', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Member Create', 'projectMembers.create'),
('57ead3c5-b09f-4cd8-ac38-d9c0d4654af5', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Member Read', 'projectMembers.read'),
('58719469-e01d-4667-bafa-12020931e317', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Member Edit', 'projectMembers.edit'),
('1a1303f3-a929-41b3-be7f-d372e6aab87a', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Member Delete', 'projectMembers.delete'),
('819adaec-f4a2-4438-937d-fcd9cc15d76b', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Work Unit Create', 'projectWorkUnits.create'),
('7c2d5fef-3096-4b68-bd76-f26cb2a3baea', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Work Unit Read', 'projectWorkUnits.read'),
('cb90c56a-60a8-4345-8001-bb7ab172b302', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Work Unit Edit', 'projectWorkUnits.edit'),
('6eb3bf29-87dc-490b-85b4-e591d416ac8f', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Work Unit Delete', 'projectWorkUnits.delete'),
('615cf3a6-89cb-4200-81b4-47542ae8b145', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Feedback Create', 'feedbacks.create'),
('cc88051f-7eea-468e-a9ac-91757dde2581', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Feedback Read', 'feedbacks.read'),
('3b542878-f597-421f-b4ac-5f71b0e22ddb', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Feedback Edit', 'feedbacks.edit'),
('2969fd2a-ae44-414a-85ac-7a930c6de987', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Feedback Delete', 'feedbacks.delete'),
('37219e6a-34f1-4d68-8727-db90ddd6f97e', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Event Question Create', 'employeeEventQuestions.create'),
('bfd6144d-0161-41f2-9cf5-3467d2a505fd', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Event Question Edit', 'employeeEventQuestions.edit'),
('be88150b-3ad5-4cef-8716-90873d194207', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Event Question Read', 'employeeEventQuestions.read'),
('23fe9fa0-d848-491f-8b6d-04ada45b9c51', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Event Question Delete', 'employeeEventQuestions.delete'),
('d5c1142a-86a6-4347-bdef-2ad41492b738', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Survey Create', 'surveys.create'),
('ef32d604-42ec-40f8-9c34-08431d2c20d8', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Survey Read', 'surveys.read'),
('f5906795-60d1-4e5f-88f2-e15f2fae2327', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Survey Edit', 'surveys.edit'),
('dfc32d81-abfa-49ce-a8f6-cdc83c8da78b', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Survey Delete', 'surveys.delete');

INSERT INTO public.role_permissions (id, deleted_at, created_at, updated_at, role_id, permission_id) VALUES
('75012959-344d-449f-a7ec-00023d68b32b', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', 'd796884d-a8c4-4525-81e7-54a3b6099eac', '01ed1076-4028-4fdf-9a92-cb57a8e041af'),
('f73306cc-5b01-49bb-87af-a2c2af14bfae', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '01ed1076-4028-4fdf-9a92-cb57a8e041af'),
('ef3e5ee2-69b5-44a6-85fd-0b8e5329ae6b', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'a7549f30-987c-47d1-9266-452f3cfc68b7'),
('6ad4a140-86ee-442b-8a7a-a0566dbab1f4', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'e4e68398-da67-438f-9fb0-07a779b504a0'),
('531e689e-9be2-442b-ac31-111f12e11a89', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '33279d12-6daa-41d3-a037-9f805e8ebf61'),
('29a63fdc-136d-4f5b-8103-32a16f3dc565', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '09a5fa4c-ad07-4dec-a16d-34e0f567ef1d'),
('b3c345fd-e800-4a64-9f8c-56158e5918c5', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '35bb795a-0b9f-428c-861d-48c1e8d4e73a'),
('4a5fd07a-e308-4032-9a73-ace558e043c6', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '80a2c800-02ea-4264-9289-57b92e911097'),
('a7b14563-51a8-4df9-a976-830ee2b12d47', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '6797aeac-7d5e-4216-b260-a8a3f62d3cf6'),
('d08730ac-c3ff-44b9-983c-9effda8284ac', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '57ead3c5-b09f-4cd8-ac38-d9c0d4654af5'),
('e2558cc9-73cf-4f60-be90-2a39ee2e95d1', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '58719469-e01d-4667-bafa-12020931e317'),
('4d5ab598-e4e9-43ed-9b26-a4f42454c46d', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '1a1303f3-a929-41b3-be7f-d372e6aab87a'),
('e242544a-24da-4820-8d28-8aa67001033d', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '819adaec-f4a2-4438-937d-fcd9cc15d76b'),
('1aef6c68-cc6c-4db3-b78f-c40b714ca4aa', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '7c2d5fef-3096-4b68-bd76-f26cb2a3baea'),
('62256fae-d6af-4246-807f-83cd580b1ad3', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'cb90c56a-60a8-4345-8001-bb7ab172b302'),
('05043830-fb1a-4558-9b5c-be9bea090cad', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '6eb3bf29-87dc-490b-85b4-e591d416ac8f'),
('8419b6c4-c09a-4dc9-ad0a-73dad343b1f0', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '615cf3a6-89cb-4200-81b4-47542ae8b145'),
('659369d5-577a-444f-b4c9-fc9c3b3bae2d', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'cc88051f-7eea-468e-a9ac-91757dde2581'),
('24783453-3676-4419-b46c-bbfff792bfda', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '3b542878-f597-421f-b4ac-5f71b0e22ddb'),
('71dffa83-cc11-4dad-a63e-4ef750941e8a', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '2969fd2a-ae44-414a-85ac-7a930c6de987'),
('b10b076f-c7bc-4c1a-a591-50baf56e125b', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '37219e6a-34f1-4d68-8727-db90ddd6f97e'),
('57b6a898-1e39-4bf5-b2ea-132d10455202', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'bfd6144d-0161-41f2-9cf5-3467d2a505fd'),
('76d22f5c-8301-4e4a-b50a-da6a7d663ba0', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'be88150b-3ad5-4cef-8716-90873d194207'),
('917e223d-dd4c-4a4c-be5c-72542e172878', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '23fe9fa0-d848-491f-8b6d-04ada45b9c51'),
('3f8e359e-610f-430a-be64-355ae8c15161', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'd5c1142a-86a6-4347-bdef-2ad41492b738'),
('f949284f-5c23-40c9-bdc9-5987e5301c50', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'ef32d604-42ec-40f8-9c34-08431d2c20d8'),
('263e57fc-f171-4a6f-a285-2b7d538ecec6', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'f5906795-60d1-4e5f-88f2-e15f2fae2327'),
('d1b42bd7-32f1-4f14-8dcc-009375ae196f', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'dfc32d81-abfa-49ce-a8f6-cdc83c8da78b');

INSERT INTO public.positions (id, deleted_at, created_at, updated_at, name, code) VALUES
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Frontend', 'frontend'),
('d796884d-a8c4-4525-81e7-54a3b6099eac', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Backend', 'backend'),
('dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Devops', 'devops'),
('01fb6322-d727-47e3-a242-5039ea4732fc', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Blockchain', 'blockchain'),
('39735742-829b-47f3-8f9d-daf0983914e5', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Project-Management', 'project-management');

INSERT INTO public.countries (id, deleted_at, created_at, updated_at, name, code, cities) VALUES
('4ef64490-c906-4192-a7f9-d2221dadfe4c',NULL,'2022-11-08 08:06:56.068148','2022-11-08 08:06:56.068148','Vietnam','+84','["Hồ Chí Minh", "An Giang", "Bà Rịa-Vũng Tàu", "Bình Dương", "Bình Định", "Bình Phước", "Bình Thuận", "Bạc Liêu", "Bắc Giang", "Bắc Kạn", "Bắc Ninh", "Bến Tre", "Cao Bằng", "Cà Mau", "Cần Thơ", "Điện Biên", "Đà Nẵng", "Đắk Lắk", "Đồng Nai", "Đắk Nông", "Đồng Tháp", "Gia Lai", "Hoà Bình", "Hà Giang", "Hà Nam", "Hà Nội", "Hà Tĩnh", "Hải Dương", "Hải Phòng", "Hậu Giang", "Hưng Yên", "Khánh Hòa", "Kiên Giang", "Kon Tum", "Lai Châu", "Lâm Đồng", "Lạng Sơn", "Lào Cai", "Long An", "Nam Định", "Nghệ An", "Ninh Bình", "Ninh Thuận", "Phú Thọ", "Phú Yên", "Quảng Bình", "Quảng Nam", "Quảng Ngãi", "Quảng Ninh", "Quảng Trị", "Sóc Trăng", "Sơn La", "Thanh Hóa", "Thái Bình", "Thái Nguyên", "Thừa Thiên Huế", "Tiền Giang", "Trà Vinh", "Tuyên Quang", "Tây Ninh", "Vĩnh Long", "Vĩnh Phúc", "Yên Bái"]'),
('da9031ce-0d6e-4344-b97a-a2c44c66153e',NULL,'2022-11-08 08:08:09.881727','2022-11-08 08:08:09.881727','Singapore','+65','["Singapore"]');

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

INSERT INTO public.chapters (id, deleted_at, created_at, updated_at, name, code, lead_id) VALUES
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Web', 'web',NULL),
('d796884d-a8c4-4525-81e7-54a3b6099eac', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Backend', 'backend',NULL),
('dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'BlockChain', 'blockchain',NULL),
('01fb6322-d727-47e3-a242-5039ea4732fc', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'DevOps', 'devops',NULL),
('01fb6322-d727-47e3-a242-5039ea4732fd', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Mobile', 'mobile',NULL),
('01fb6322-d727-47e3-a242-5039ea4732fe', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'QA', 'qa',NULL),
('39735742-829b-47f3-8f9d-daf0983914e5', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'PM', 'pm',NULL);

INSERT INTO public.currencies (id, deleted_at, created_at, updated_at, name, symbol, locale, type) VALUES
('06a699ed-618b-400b-ac8c-8739956fa8e7', NULL, '2019-02-20 04:24:14.967209+00', '2019-02-20 04:24:14.967209+00', 'GBP', '£', 'en-gb', 'fiat'),
('0a6f4a2e-a097-4f7e-ae65-bfee3298e5cc', NULL, '2020-11-23 08:59:49.802559+00', '2020-11-23 08:59:49.802559+00', 'TEN', NULL, NULL, 'crypto'),
('1c7dcbe2-6984-461d-8ed9-537676f2b590', NULL, '2021-07-21 18:10:24.194492+00', '2021-07-21 18:10:24.194492+00', 'USD', '$', 'en-us', 'crypto'),
('283e559f-a7e7-4aa7-9e08-4806d6c07016', NULL, '2019-02-20 04:24:14.967209+00', '2019-02-20 04:24:14.967209+00', 'EUR', '€', 'en-gb', 'fiat'),
('2de81125-a947-4fea-a006-2c60e7ec01ed', NULL, '2020-02-28 07:18:02.349700+00', '2020-02-28 07:18:02.349700+00', 'CAD', 'c$', 'en-ca', 'fiat'),
('7037bdb6-584e-4e35-996d-ef28a243f48a', NULL, '2019-02-11 06:14:51.305496+00', '2019-02-11 06:14:51.305496+00', 'VND', 'đ', 'vi-vn', 'fiat'),
('bf256e69-28b0-4d9f-bf48-3662854157a9', NULL, '2019-10-28 14:14:52.302051+00', '2019-10-28 14:14:52.302051+00', 'SGD', 's$', 'en-sg', 'fiat'),
('f00498e4-7a4c-4f61-b126-b84b5faeee06', NULL, '2019-02-11 06:14:51.305496+00', '2019-02-11 06:14:51.305496+00', 'USD', '$', 'en-us', 'fiat');

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
