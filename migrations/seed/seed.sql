INSERT INTO public.roles (id, deleted_at, created_at, updated_at, name, code) VALUES
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Admin', 'admin'),
('d796884d-a8c4-4525-81e7-54a3b6099eac', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Member', 'member');

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
('37219e6a-34f1-4d68-8727-db90ddd6f97e', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Event Question Read', 'employeeEventQuestions.read');

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
('71dffa83-cc11-4dad-a63e-4ef750941e8a', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '2969fd2a-ae44-414a-85ac-7a930c6de987');
('b10b076f-c7bc-4c1a-a591-50baf56e125b', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '37219e6a-34f1-4d68-8727-db90ddd6f97e');

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


INSERT INTO public.employees (id,deleted_at,created_at,updated_at,full_name,display_name,gender,team_email,personal_email,avatar,phone_number,address,mbti,horoscope,passport_photo_front,passport_photo_back,identity_card_photo_front,identity_card_photo_back,date_of_birth,"working_status",joined_date,left_date,basecamp_id,basecamp_attachable_sgid,gitlab_id,github_id,discord_id,wise_recipient_email,wise_recipient_name,wise_recipient_id,wise_account_number,wise_currency,local_bank_branch,local_bank_number,local_bank_currency,local_branch_name,local_bank_recipient_name,position_id,seniority_id,notion_id,line_manager_id) VALUES
('f7c6016b-85b5-47f7-8027-23c2db482197', NULL, '2022-11-17 11:06:03.275039', '2022-11-17 11:06:03.275039', 'Nguyen Van Duc','Duc Nguyen','Male','ducnv@dwarvesv.com','nguyenduc123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/7227222220023148542.png','0123456789','Hoa Khanh Bac - Da Nang - Viet Nam','INFP-A','Libra',NULL,NULL,NULL,NULL,'2000-02-02','probation','2022-05-25',NULL,NULL,NULL,'ducnv','ducnv','236885900779454465',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'d796884d-a8c4-4525-81e7-54a3b6099eac','26581472','2655832e-f009-4b73-a535-64c3a22e558f'),
('d42a6fca-d3b8-4a48-80f7-a95772abda56', NULL, '2022-11-17 11:06:03.279339', '2022-11-17 11:06:03.279339', 'Nguyễn Xuân Trường','Truong Nguyen','Male','truongnx@dwarvesv.com','nguyentruong123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/8690187460973853786.png','0123456789','Nam Tu Liem','ISFJ-T','Virgo',NULL,NULL,NULL,NULL,'1995-04-05','full-time','2022-05-25',NULL,NULL,NULL,'truongnguyen','truongnguyen','205167514731151360',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'11ccffea-2cc9-4e98-9bef-3464dfe4dec8','26581472','2655832e-f009-4b73-a535-64c3a22e558f'),
('dcfee24b-306d-4609-9c24-a4021639a11b', NULL, '2022-11-17 11:06:03.264371', '2022-11-17 11:06:03.264371', 'Lê Việt Quỳnh','Quynh Le','Female','quynhlv@dwarvesv.com','lequynh123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/4916620003720295041.png','0123456789','Binh Tan District','INTJ-S','Virgo',NULL,NULL,NULL,NULL,'1998-12-10','full-time','2021-05-17',NULL,NULL,NULL,'quynhle','quynhle','435844136025849876',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'01fb6322-d727-47e3-a242-5039ea4732fc','29088916','2655832e-f009-4b73-a535-64c3a22e558f'),
('38a00d4a-bc45-41de-965a-adc674ab82c9', NULL, '2022-11-17 11:15:40.397372', '2022-11-17 11:15:40.397372', 'Nguyễn Hữu Nguyên','Nguyen Nguyen','Male','nguyennh@dwarvesv.com','nguyennh123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/2137109913829673789.png','0123456789','Q. Thu Duc , TP.HCM','ISTJ-A','Libra',NULL,NULL,NULL,NULL,'1995-04-05','probation','2022-05-25',NULL,NULL,NULL,'nguyennh','nguyennh','205167514731151360',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '26581472','2655832e-f009-4b73-a535-64c3a22e558f'),
('ecea9d15-05ba-4a4e-9787-54210e3b98ce', NULL, '2022-11-02 09:50:55.320669', '2022-11-02 09:50:55.320669', 'Nguyễn Hoàng Huy','Huy Nguyen','Male','huynh@d.foundation','hoanghuy123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/2830497479497502617.png','0123456789','Thu Duc City, Ho Chi Minh City','ENFJ-A','Aquaman',NULL,NULL,NULL,NULL,'1990-01-02','probation','2018-09-01',NULL,NULL,NULL,'huynguyen','huynguyen','646649040771219476',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'01fb6322-d727-47e3-a242-5039ea4732fd','39523061','2655832e-f009-4b73-a535-64c3a22e558f'),
('2655832e-f009-4b73-a535-64c3a22e558f', NULL, '2022-11-02 09:52:34.586566', '2022-11-02 09:52:34.586566', 'Phạm Đức Thành','Thanh Pham','Male','thanh@d.foundation','thanhpham123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/5153574695663955944.png','0123456788','Tan Binh District, Ho Chi Minh, Vietnam','ISFJ-A','Aquaman',NULL,NULL,NULL,NULL,'1990-01-02','contractor','2018-09-01',NULL,NULL,NULL,'thanhpham','thanhpham','646649040771219476',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'01fb6322-d727-47e3-a242-5039ea4732fd','39523061',NULL),
('608ea227-45a5-4c8a-af43-6c7280d96340', NULL, '2022-11-02 09:50:55.320669', '2022-11-02 09:50:55.320669', 'Tiêu Quang Huy','Huy Tieu','Male','huytq@d.foundation','huytq123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/3c420751-bb9a-4878-896e-2f10f3a633d6_avatar2535921977139052349.png','0123456789','Binh Thanh, Ho Chi Minh City','INFP-A','Libra',NULL,NULL,NULL,NULL,'1990-01-02','contractor','2018-09-01',NULL,NULL,NULL,'huytq','huytq','435844136025849876',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'01fb6322-d727-47e3-a242-5039ea4732fd','39523061','2655832e-f009-4b73-a535-64c3a22e558f'),
('3f705527-0455-4e67-a585-6c1f23726fff', NULL, '2022-11-17 11:06:03.245183', '2022-11-17 11:06:03.245183', 'Giang Ngọc Huy','Huy Giang','Male','huy@dwarvesv.com','gnhuy123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/261887286241984054.png','0123456789','Binh Tan District, Ho Chi Minh, Vietnam','ISFJ-A','Libra',NULL,NULL,NULL,NULL,'1991-11-11','full-time','2018-09-01',NULL,NULL,NULL,'huygn','huygn','435844136025849876',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'01fb6322-d727-47e3-a242-5039ea4732fd','29088916','2655832e-f009-4b73-a535-64c3a22e558f'),
('37e00d47-de69-4ac8-991b-cf3e39565c00', NULL, '2022-11-17 11:06:03.262847', '2022-11-17 11:06:03.262847', 'Phùng Thị Thương Thương','Thuong Phung','Female','thuongptt@dwarvesv.com','thuongphung123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/7186955696154477566.png','0123456789','District 10, Ho Chi Minh, Vietnam','INFP-A','Scorpio',NULL,NULL,NULL,NULL,'1991-11-11','full-time','2021-05-17',NULL,NULL,NULL,'thuongphuong','thuongphuong','435844136025849876',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'01fb6322-d727-47e3-a242-5039ea4732fc', '29088916','2655832e-f009-4b73-a535-64c3a22e558f'),
('fae443f8-e8ff-4eec-b86c-98216d7662d8', NULL, '2022-11-17 11:06:03.270325', '2022-11-17 11:06:03.270325', 'Ho Quang Toan','Toan Ho','Male','toanhq@dwarvesv.com','toanbku123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/6674226719566030792.png','0123456789','Tam Phu ward, Thu Duc district','ISFJ-A','Libra',NULL,NULL,NULL,NULL,'1998-12-10','left','2022-05-25','2022-04-12',NULL,NULL,'toanhq','toanhq','236885900779454465',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e','26581472','2655832e-f009-4b73-a535-64c3a22e558f'),
('ac318f73-0c8e-43ed-b00e-d230670dc400', NULL, '2022-11-17 11:06:03.271908', '2022-11-17 11:06:03.271908', 'Trần Diễm Quỳnh','Quynh Tran','Female','quynhtd@dwarvesv.com','quynhtd123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/8900807301346129018.png','0123456789','Tay Ho','ISTJ-A','Aries',NULL,NULL,NULL,NULL,'2000-02-02','left','2022-05-25','2022-04-12',NULL,NULL,'quynhtd','quynhtd','236885900779454465',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', '26581472','2655832e-f009-4b73-a535-64c3a22e558f'),
('d389d35e-c548-42cf-9f29-2a599969a8f2', NULL, '2022-11-17 11:06:03.267167', '2022-11-17 11:06:03.267167', 'Nguyễn Trần Minh Thảo','Thao Nguyen','Female','thaontm@dwarvesv.com','thaonguyen123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/5348640939694049226.png','0123456789','Bình Tân District','INFP-A','Virgo',NULL,NULL,NULL,NULL,'1998-12-10','full-time','2021-05-17',NULL,NULL,NULL,'thaonguyen','thaonguyen','435844136025849876',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', '29088916','2655832e-f009-4b73-a535-64c3a22e558f'),
('a1f25e3e-cf40-4d97-a4d5-c27ee566b8c5', NULL, '2022-11-17 11:06:03.265726', '2022-11-17 11:06:03.265726', 'Phan Duy Cường','Cuong Phan','Male','cody@dwarvesv.com','cuongphan123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/8465254073957996724.png','0123456789','Tan Binh Ward','ISFJ-A','Virgo',NULL,NULL,NULL,NULL,'1998-12-10','full-time','2021-05-17',NULL,NULL,NULL,'cuongphan','cuongphan','435844136025849876',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', '29088916','2655832e-f009-4b73-a535-64c3a22e558f'),
('498d5805-dd64-4643-902d-95067d6e5ab5', NULL, '2022-11-17 11:06:03.268733', '2022-11-17 11:06:03.268733', 'Võ Hải Biên','Bien Vo','Male','bean@dwarvesv.com','vhbien123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/4114345912835654270.png','0123456789','Quận Thủ Đức - TPHCM','ISFJ-T','Libra',NULL,NULL,NULL,NULL,'1998-12-10','full-time','2021-05-17',NULL,NULL,NULL,'vhbien','vhbien','236885900779454465',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', '29088916','2655832e-f009-4b73-a535-64c3a22e558f'),
('f6ce0d0f-5794-463b-ad0b-8240ab9c49be', NULL, '2022-11-17 11:06:03.276614', '2022-11-17 11:06:03.276614', 'Nguyễn Hoàng Anh','Anh Nguyen','Male','anhnh@dwarvesv.com','nguyenanh123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/3274097965772922906.png','0123456789','Q.Tân Bình, TP.HCM','INFP-A','Aries',NULL,NULL,NULL,NULL,'2000-02-02','full-time','2022-05-25',NULL,NULL,NULL,'anhnh','anhnh','236885900779454465',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'d796884d-a8c4-4525-81e7-54a3b6099eac', '26581472','2655832e-f009-4b73-a535-64c3a22e558f'),
('061820c0-bf6c-4b4a-9753-875f75d71a2c', NULL, '2022-11-17 11:06:03.278004', '2022-11-17 11:06:03.278004', 'Vong Tieu Hung','Hung Vong','Male','hungvt@dwarvesv.com','tieuhung123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/5762098830282049280.png','0123456789','Thu Duc','INFP-A','Aries',NULL,NULL,NULL,NULL,'1995-04-05','probation','2022-05-25',NULL,NULL,NULL,'hungvong','hungvong','205167514731151360',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'d796884d-a8c4-4525-81e7-54a3b6099eac', '26581472','2655832e-f009-4b73-a535-64c3a22e558f'),
('d675dfc5-acbe-4566-acde-f7cb132c0206', NULL, '2022-11-17 11:06:03.281048', '2022-11-17 11:06:03.281048', 'Le Nguyen An Khang','Khang Le','Male','khanglna@dwarvesv.com','khangle123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/4368900905892223171.png','0123456789','Binh Thanh District','ENFJ-A','Libra',NULL,NULL,NULL,NULL,'1995-04-05','probation','2022-05-25',NULL,NULL,NULL,'khangle','khangle','205167514731151360',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '26581472','2655832e-f009-4b73-a535-64c3a22e558f'),
('8d7c99c0-3253-4286-93a9-e7554cb327ef', NULL, '2022-11-02 09:50:55.320669', '2022-11-02 09:50:55.320669', 'Nguyễn Hải Nam','Nam Nguyen','Male','benjamin@d.foundation','nam123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/8399103964540935617.png','0123456789','District 7, Ho Chi Minh City','ENFJ-A','Libra',NULL,NULL,NULL,NULL,'1990-01-02','probation','2018-09-01',NULL,NULL,NULL,'namnh','namnh','646649040771219476',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'01fb6322-d727-47e3-a242-5039ea4732fd','39523061','2655832e-f009-4b73-a535-64c3a22e558f'),
('eeae589a-94e3-49ac-a94c-fcfb084152b2', NULL, '2022-11-02 09:50:55.320669', '2022-11-02 09:50:55.320669', 'Nguyễn Ngô Lập','Lap Nguyen','Male','alan@d.foundation','lap123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/2870969541970972723.png','0123456789','Binh chanh, Ho Chi Minh City','ENFJ-A','Libra',NULL,NULL,NULL,NULL,'1990-01-02','probation','2018-09-01',NULL,NULL,NULL,'lapnn','lapnn','646649040771219476',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'01fb6322-d727-47e3-a242-5039ea4732fd','39523061','2655832e-f009-4b73-a535-64c3a22e558f'),
('d1b1dcbe-f4ed-49cc-a16b-56c9a3145c14', NULL, '2022-11-17 11:06:03.255694', '2022-11-17 11:06:03.255694', 'Nghiêm Minh Đức','Duc Nghiem','Male','duc@dwarvesv.com','ducnghiem123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/3851063061598282120.png','0123456789','Tan Binh District, Ho Chi Minh, Vietnam','INTJ-S','Scorpio',NULL,NULL,NULL,NULL,'1991-11-11','probation','2021-05-17',NULL,NULL,NULL,'duc','duc','435844136025849876',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'01fb6322-d727-47e3-a242-5039ea4732fc','29088916','2655832e-f009-4b73-a535-64c3a22e558f'),
('7bcf4b45-0279-4da2-84e4-eec5d9d05ba3', NULL, '2022-11-17 11:06:03.261104', '2022-11-17 11:06:03.261104', 'Nguyễn Lâm Ngọc Duy','Duy Nguyen','Female','mia@dwarvesv.com','ngocduy123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/7186955696154477566.png','0123456789','Binh An, District 2, Ho Chi Minh, Vietnam','INTJ-S','Scorpio',NULL,NULL,NULL,NULL,'1991-11-11','full-time','2021-05-17',NULL,NULL,NULL,'mia','mia','435844136025849876',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'01fb6322-d727-47e3-a242-5039ea4732fc','29088916','2655832e-f009-4b73-a535-64c3a22e558f'),
('7fbfb59b-e00e-46b2-85cd-64f9f9942daa', NULL, '2022-11-17 11:06:03.27344', '2022-11-17 11:06:03.27344', 'Đoàn Tấn Đạt','Dat Doan','Male','datdt@dwarvesv.com','doandat123@gmail.com','https://s3-ap-southeast-1.amazonaws.com/fortress-images/1105085008506574007.png','0123456789','Binh Tan District, Ho Chi Minh, Vietnam','INFP-A','Aries',NULL,NULL,NULL,NULL,'2000-02-02','left','2022-05-25','2022-04-12',NULL,NULL,'datdt','datdt','236885900779454465',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'d796884d-a8c4-4525-81e7-54a3b6099eac','26581472','2655832e-f009-4b73-a535-64c3a22e558f');

INSERT INTO public.stacks (id, deleted_at, created_at, updated_at, name, code, avatar) VALUES
('0ecf47c8-cca4-4c30-94bb-054b1124c44f', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'Golang', 'golang', NULL),
('fa0f4e46-7eab-4e5c-9d31-30489e69fe2e', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'React', 'react', NULL),
('b403ef95-4269-4830-bbb6-8e56e5ec0af4', NULL, '2022-11-11 18:38:46.266725', '2022-11-11 18:38:46.266725', 'Google Cloud', 'gcloud', NULL);

INSERT INTO public.chapters (id, deleted_at, created_at, updated_at, name, code, lead_id) VALUES
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Web', 'web','8d7c99c0-3253-4286-93a9-e7554cb327ef'),
('d796884d-a8c4-4525-81e7-54a3b6099eac', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Backend', 'backend',NULL),
('dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'BlockChain', 'blockchain',NULL),
('01fb6322-d727-47e3-a242-5039ea4732fc', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'DevOps', 'devops',NULL),
('01fb6322-d727-47e3-a242-5039ea4732fd', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Mobile', 'mobile',NULL),
('01fb6322-d727-47e3-a242-5039ea4732fe', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'QA', 'qa',NULL),
('39735742-829b-47f3-8f9d-daf0983914e5', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'PM', 'pm',NULL);

INSERT INTO public.employee_roles (id, deleted_at, created_at, updated_at, employee_id, role_id) VALUES
('f40ead54-dc57-40d5-9c6a-4f22341ee7d2', NULL, '2022-11-11 18:36:31.657696', '2022-11-11 18:36:31.657696', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('3fd72903-cebb-4519-82c2-4c8f096a272b', NULL, '2022-11-11 18:36:31.657696', '2022-11-11 18:36:31.657696', 'eeae589a-94e3-49ac-a94c-fcfb084152b2', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('b3c27435-eb39-400b-a94b-7bf8218b8b3e', NULL, '2022-11-11 18:36:31.657696', '2022-11-11 18:36:31.657696', '2655832e-f009-4b73-a535-64c3a22e558f', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('5bfea659-4912-4668-ae28-644083595aa6', NULL, '2022-11-11 18:36:31.657696', '2022-11-11 18:36:31.657696', '608ea227-45a5-4c8a-af43-6c7280d96340', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('178a7b5c-ddac-445d-9045-09d3b6b04431', NULL, '2022-11-11 18:36:31.657696', '2022-11-11 18:36:31.657696', '8d7c99c0-3253-4286-93a9-e7554cb327ef', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('eb45824b-38c0-4ca9-84c0-a9e3cd9020e3', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', '3f705527-0455-4e67-a585-6c1f23726fff', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('12bf24b0-b95a-4be4-8cf8-c57e9058a278', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', 'd1b1dcbe-f4ed-49cc-a16b-56c9a3145c14', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('37282ed4-01a7-4cdc-af50-e5393e957e14', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', '7bcf4b45-0279-4da2-84e4-eec5d9d05ba3', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('bd9500a6-29b9-4d00-a326-22983d10d834', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', '37e00d47-de69-4ac8-991b-cf3e39565c00', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('29c3976d-f937-4acc-956b-0bc61011cea5', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', 'dcfee24b-306d-4609-9c24-a4021639a11b', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('4a819c90-0861-4902-9f73-520727ac01bd', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', 'a1f25e3e-cf40-4d97-a4d5-c27ee566b8c5', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('61dfa7c3-4b1f-43af-960e-e4a4080b37de', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', 'd389d35e-c548-42cf-9f29-2a599969a8f2', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('35e93f5c-9a33-49cc-ac3a-d05d3cc56532', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', '498d5805-dd64-4643-902d-95067d6e5ab5', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('2afab2e6-e955-4b6c-9231-be378196acf1', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', 'fae443f8-e8ff-4eec-b86c-98216d7662d8', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('042425e7-a53a-4945-995b-a878014e4c29', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', 'ac318f73-0c8e-43ed-b00e-d230670dc400', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('b8760efa-eaaf-434a-9ed3-c2bab9b013ee', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', '7fbfb59b-e00e-46b2-85cd-64f9f9942daa', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('cddf3877-f0ff-4274-bb2e-cd55c00474b2', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', 'f7c6016b-85b5-47f7-8027-23c2db482197', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('87917065-69aa-48f8-9afb-aa166b938d98', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', 'f6ce0d0f-5794-463b-ad0b-8240ab9c49be', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('8a188594-09bb-490b-8c16-90220469f477', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', '061820c0-bf6c-4b4a-9753-875f75d71a2c', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('436cd54e-7f3a-4ab7-a581-b0aad91a7807', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', 'd42a6fca-d3b8-4a48-80f7-a95772abda56', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('e8c8619d-295b-4d55-936f-cd6406f0857f', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', 'd675dfc5-acbe-4566-acde-f7cb132c0206', 'd796884d-a8c4-4525-81e7-54a3b6099eac'),
('f7a25073-6832-4ba0-a69e-b0a9c87c61b4', NULL, '2022-11-17 11:53:14.690017', '2022-11-17 11:53:14.690017', '38a00d4a-bc45-41de-965a-adc674ab82c9', 'd796884d-a8c4-4525-81e7-54a3b6099eac');

INSERT INTO public.projects (id, deleted_at, created_at, updated_at, name, type, start_date, end_date, status, country_id, client_email, project_email) VALUES
('dfa182fc-1d2d-49f6-a877-c01da9ce4207', NULL, '2022-11-11 18:06:56.362902', '2022-11-11 18:06:56.362902', 'Lorem ipsum', 'time-material', '2022-07-06', NULL, 'active', NULL, NULL, NULL),
('8dc3be2e-19a4-4942-8a79-56db391a0b15', NULL, '2022-11-11 18:06:56.362902', '2022-11-11 18:06:56.362902', 'Fortress', 'dwarves', '2022-11-01', NULL, 'active', '4ef64490-c906-4192-a7f9-d2221dadfe4c', 'team@d.foundation', 'fortress@d.foundation');

INSERT INTO public.project_slots (id, deleted_at, created_at, updated_at, project_id, seniority_id, upsell_person_id, position, deployment_type, rate, discount, status) VALUES
('f32d08ca-8863-4ab3-8c84-a11849451eb7', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL,'frontend', 'official', 5000, 0, 'active'),
('bdc64b18-4c5f-4025-827a-f5b91d599dc7', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL,'backend', 'shadow', 4000, 0, 'active'),
('1406fcce-6f90-4e0f-bea1-c373e2b2b5b1', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL,'backend', 'official', 3000, 3, 'active'),
('b25bd3fa-eb6d-49d5-b278-7aacf4594f79', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL,'frontend', 'official', 3000, 3, 'active'),
('ce379dc0-95be-471a-9227-8e045a5630af', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL,'project-management', 'shadow', 4000, 0, 'active');

INSERT INTO public.project_members (id, deleted_at, created_at, updated_at, project_id, project_slot_id ,employee_id, seniority_id, joined_date, left_date, rate, discount, status,deployment_type, upsell_person_id) VALUES
('cb889a9c-b20c-47ee-83b8-44b6d1721aca', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'f32d08ca-8863-4ab3-8c84-a11849451eb7', '2655832e-f009-4b73-a535-64c3a22e558f', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', NULL, 5000, 0, 'active', 'official', '608ea227-45a5-4c8a-af43-6c7280d96340'),
('7310b51a-3855-498b-99ab-41aa82934269', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'bdc64b18-4c5f-4025-827a-f5b91d599dc7', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', NULL, 4000, 0, 'active', 'shadow', '608ea227-45a5-4c8a-af43-6c7280d96340'),
('35149aab-0506-4eb3-9300-c706ccbf2bde', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '1406fcce-6f90-4e0f-bea1-c373e2b2b5b1', '8d7c99c0-3253-4286-93a9-e7554cb327ef', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', NULL, 3000, 3, 'active', 'official', '608ea227-45a5-4c8a-af43-6c7280d96340'),
('5a9a07aa-e8f3-4b62-b9ad-0f057866dc6c', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'b25bd3fa-eb6d-49d5-b278-7aacf4594f79', 'eeae589a-94e3-49ac-a94c-fcfb084152b2', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', NULL, 3000, 3, 'active', 'official', '608ea227-45a5-4c8a-af43-6c7280d96340'),
('fcd5c16f-40fd-48b6-9649-410691373eea', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'ce379dc0-95be-471a-9227-8e045a5630af', '608ea227-45a5-4c8a-af43-6c7280d96340', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', NULL, 4000, 0, 'active', 'shadow', '608ea227-45a5-4c8a-af43-6c7280d96340');

INSERT INTO public.project_heads (id, deleted_at, created_at, updated_at, project_id, employee_id, joined_date, left_date, commission_rate, position) VALUES
('528433a5-4001-408d-bd0c-3b032eef0a70', NULL, '2022-11-11 18:27:03.906845', '2022-11-11 18:27:03.906845', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '2655832e-f009-4b73-a535-64c3a22e558f', '2022-11-01', NULL, 1, 'account-manager'),
('14d1d9f0-5f0f-49d5-8309-ac0a40de013d', NULL, '2022-11-11 18:27:03.906845', '2022-11-11 18:27:03.906845', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '8d7c99c0-3253-4286-93a9-e7554cb327ef', '2022-11-11', NULL, 1, 'technical-lead'),
('e17b6bfc-79b0-4d65-ac88-559d8c597d2e', NULL, '2022-11-11 18:27:03.906845', '2022-11-11 18:27:03.906845', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '608ea227-45a5-4c8a-af43-6c7280d96340', '2022-11-02', NULL, 2, 'sale-person'),
('72578fb2-921b-4c8b-b998-a4aec73da809', NULL, '2022-11-11 18:27:03.906845', '2022-11-11 18:27:03.906845', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', '2022-11-01', '2022-11-11', 2.5, 'technical-lead'),
('eca9fed2-21cf-4c75-9ab5-f6e5a13a0a75', NULL, '2022-11-11 18:27:03.906845', '2022-11-11 18:27:03.906845', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '2655832e-f009-4b73-a535-64c3a22e558f', '2022-11-01', NULL, 0.5, 'delivery-manager');

INSERT INTO public.project_stacks (id, deleted_at, created_at, updated_at, project_id, stack_id) VALUES
('6b89cc06-33ca-4f48-825a-ab20da5cb287', NULL, '2022-11-11 18:39:34.923619', '2022-11-11 18:39:34.923619', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('da982888-1727-4c06-8e8a-86aff1f80f89', NULL, '2022-11-11 18:39:34.923619', '2022-11-11 18:39:34.923619', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'fa0f4e46-7eab-4e5c-9d31-30489e69fe2e'),
('371d797f-a639-4869-8026-67e682c6da4a', NULL, '2022-11-11 18:39:34.923619', '2022-11-11 18:39:34.923619', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'b403ef95-4269-4830-bbb6-8e56e5ec0af4');

INSERT INTO public.employee_stacks (id,deleted_at,created_at,updated_at,employee_id,stack_id) VALUES
('a9921fe8-3b77-46ad-a76d-172cb85e10a5',NULL,'2022-11-16 11:28:01.048359','2022-11-16 11:28:01.048359','ecea9d15-05ba-4a4e-9787-54210e3b98ce','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('94de566e-9d1c-4ce7-a16a-0deaf3a45f9a',NULL,'2022-11-16 11:28:01.051869','2022-11-16 11:28:01.051869','ecea9d15-05ba-4a4e-9787-54210e3b98ce','fa0f4e46-7eab-4e5c-9d31-30489e69fe2e'),
('8eb2ffa5-68aa-4b0c-b113-e8b37297031e',NULL,'2022-11-16 11:28:01.053014','2022-11-16 11:28:01.053014','2655832e-f009-4b73-a535-64c3a22e558f','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('e2c3d67b-47e3-4020-8684-2971b3e12b9b',NULL,'2022-11-16 11:28:01.054087','2022-11-16 11:28:01.054087','2655832e-f009-4b73-a535-64c3a22e558f','b403ef95-4269-4830-bbb6-8e56e5ec0af4'),
('c3fd413f-1a7e-4be9-939b-97745ae98715',NULL,'2022-11-16 11:28:01.059009','2022-11-16 11:28:01.059009','8d7c99c0-3253-4286-93a9-e7554cb327ef','fa0f4e46-7eab-4e5c-9d31-30489e69fe2e'),
('da9f4cb2-0230-4c28-babb-7ed7ca9d5682',NULL,'2022-11-16 11:28:01.060256','2022-11-16 11:28:01.060256','8d7c99c0-3253-4286-93a9-e7554cb327ef','b403ef95-4269-4830-bbb6-8e56e5ec0af4'),
('4c137480-98af-4dcb-a7c0-ec56cc02468b',NULL,'2022-11-16 11:28:01.06115','2022-11-16 11:28:01.06115','eeae589a-94e3-49ac-a94c-fcfb084152b2','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('37ba3955-2917-4f09-a647-74486e56b28f',NULL,'2022-11-16 11:28:01.061939','2022-11-16 11:28:01.061939','608ea227-45a5-4c8a-af43-6c7280d96340','b403ef95-4269-4830-bbb6-8e56e5ec0af4'),
('91f87f40-708e-4a6f-bda6-0ff00c243517',NULL,'2022-11-17 12:03:49.970205','2022-11-17 12:03:49.970205','3f705527-0455-4e67-a585-6c1f23726fff','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('b77ab8ab-79a4-4d89-94fe-bcc8d92af852',NULL,'2022-11-17 12:03:49.970205','2022-11-17 12:03:49.970205','d1b1dcbe-f4ed-49cc-a16b-56c9a3145c14','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('e0f3378b-12a4-4417-8ee8-93f014cf322d',NULL,'2022-11-17 12:03:49.970205','2022-11-17 12:03:49.970205','7bcf4b45-0279-4da2-84e4-eec5d9d05ba3','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('0250f55d-493f-4457-8ca5-a0033129b2ea',NULL,'2022-11-17 12:03:49.970205','2022-11-17 12:03:49.970205','37e00d47-de69-4ac8-991b-cf3e39565c00','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('3c882e45-fbf8-4eec-a66c-d69d82cbad58',NULL,'2022-11-17 12:03:49.970205','2022-11-17 12:03:49.970205','dcfee24b-306d-4609-9c24-a4021639a11b','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('ead295a9-be02-4075-90be-d22741f18663',NULL,'2022-11-17 12:03:49.970205','2022-11-17 12:03:49.970205','a1f25e3e-cf40-4d97-a4d5-c27ee566b8c5','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('79b964de-e86a-4bdd-b4ce-c946a85ddf83',NULL,'2022-11-17 12:03:49.970205','2022-11-17 12:03:49.970205','d389d35e-c548-42cf-9f29-2a599969a8f2','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('317a6519-96c5-43ed-a797-239285a2a4af',NULL,'2022-11-17 12:03:49.970205','2022-11-17 12:03:49.970205','498d5805-dd64-4643-902d-95067d6e5ab5','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('f5b88b19-0420-472a-b2e7-280ba6a87033',NULL,'2022-11-17 12:03:49.970205','2022-11-17 12:03:49.970205','fae443f8-e8ff-4eec-b86c-98216d7662d8','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('94cdb154-a8ba-4eb0-aab7-844dd9ac0150',NULL,'2022-11-17 12:03:49.970205','2022-11-17 12:03:49.970205','ac318f73-0c8e-43ed-b00e-d230670dc400','0ecf47c8-cca4-4c30-94bb-054b1124c44f'),
('72daab6f-8e68-4c0d-8ed7-4c6b1dc05000',NULL,'2022-11-17 12:04:01.64086','2022-11-17 12:04:01.64086','7fbfb59b-e00e-46b2-85cd-64f9f9942daa','fa0f4e46-7eab-4e5c-9d31-30489e69fe2e'),
('cda852a5-c08f-4358-99b4-416a16f3df0d',NULL,'2022-11-17 12:04:01.64086','2022-11-17 12:04:01.64086','f7c6016b-85b5-47f7-8027-23c2db482197','fa0f4e46-7eab-4e5c-9d31-30489e69fe2e'),
('ba971851-654c-4993-b385-99226fdc6c89',NULL,'2022-11-17 12:04:01.64086','2022-11-17 12:04:01.64086','f6ce0d0f-5794-463b-ad0b-8240ab9c49be','fa0f4e46-7eab-4e5c-9d31-30489e69fe2e'),
('6bcc6365-244c-41a6-8b8e-195be11cefa5',NULL,'2022-11-17 12:04:01.64086','2022-11-17 12:04:01.64086','061820c0-bf6c-4b4a-9753-875f75d71a2c','fa0f4e46-7eab-4e5c-9d31-30489e69fe2e'),
('a5150d8f-66d4-478c-a911-27034745fde5',NULL,'2022-11-17 12:04:01.64086','2022-11-17 12:04:01.64086','d42a6fca-d3b8-4a48-80f7-a95772abda56','fa0f4e46-7eab-4e5c-9d31-30489e69fe2e'),
('769ecb7f-59af-4227-af4c-ae95ebd8d7e9',NULL,'2022-11-17 12:04:01.64086','2022-11-17 12:04:01.64086','d675dfc5-acbe-4566-acde-f7cb132c0206','fa0f4e46-7eab-4e5c-9d31-30489e69fe2e'),
('492682dd-56ae-47c0-bd96-0616999cfc1e',NULL,'2022-11-17 12:04:01.64086','2022-11-17 12:04:01.64086','38a00d4a-bc45-41de-965a-adc674ab82c9','fa0f4e46-7eab-4e5c-9d31-30489e69fe2e'),
('38c5b4fa-ba27-4b7d-b37e-8227f1251bb9',NULL,'2022-11-17 12:04:47.072814','2022-11-17 12:04:47.072814','ecea9d15-05ba-4a4e-9787-54210e3b98ce','b403ef95-4269-4830-bbb6-8e56e5ec0af4'),
('f694cc53-7435-4b8e-b1eb-b67c28d90f48',NULL,'2022-11-17 12:04:47.072814','2022-11-17 12:04:47.072814','3f705527-0455-4e67-a585-6c1f23726fff','b403ef95-4269-4830-bbb6-8e56e5ec0af4'),
('da51e41c-f3d0-434a-9c5a-ccdd63050fc0',NULL,'2022-11-17 12:04:47.072814','2022-11-17 12:04:47.072814','d1b1dcbe-f4ed-49cc-a16b-56c9a3145c14','b403ef95-4269-4830-bbb6-8e56e5ec0af4'),
('b807dd00-1e8d-4bf3-b36a-06448ce951d3',NULL,'2022-11-17 12:04:47.072814','2022-11-17 12:04:47.072814','7bcf4b45-0279-4da2-84e4-eec5d9d05ba3','b403ef95-4269-4830-bbb6-8e56e5ec0af4'),
('3547a33b-987c-4b11-be28-922748c91fe1',NULL,'2022-11-17 12:04:47.072814','2022-11-17 12:04:47.072814','37e00d47-de69-4ac8-991b-cf3e39565c00','b403ef95-4269-4830-bbb6-8e56e5ec0af4'),
('eba10181-828f-4af4-8fc4-a974c286fd3c',NULL,'2022-11-17 12:04:47.072814','2022-11-17 12:04:47.072814','dcfee24b-306d-4609-9c24-a4021639a11b','b403ef95-4269-4830-bbb6-8e56e5ec0af4'),
('ed8ab5bf-db10-4e1a-99f4-c8a9709e3d74',NULL,'2022-11-17 12:04:47.072814','2022-11-17 12:04:47.072814','a1f25e3e-cf40-4d97-a4d5-c27ee566b8c5','b403ef95-4269-4830-bbb6-8e56e5ec0af4'),
('5907d70e-3fbb-4d83-a385-2959cc07a946',NULL,'2022-11-17 12:04:47.072814','2022-11-17 12:04:47.072814','d389d35e-c548-42cf-9f29-2a599969a8f2','b403ef95-4269-4830-bbb6-8e56e5ec0af4'),
('ab44eab9-b8e7-4ff6-bbc6-cef201c1c9dc',NULL,'2022-11-17 12:04:47.072814','2022-11-17 12:04:47.072814','498d5805-dd64-4643-902d-95067d6e5ab5','b403ef95-4269-4830-bbb6-8e56e5ec0af4');

INSERT INTO public.employee_positions (id, deleted_at, created_at, updated_at, employee_id, position_id) VALUES
('e61bf924-1eff-426f-8f71-ab49446134fb',NULL,'2022-11-16 02:45:17.982659','2022-11-16 02:45:17.982659','2655832e-f009-4b73-a535-64c3a22e558f','11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('d7b49b9e-4b51-4e3a-a7be-b94143215f7a',NULL,'2022-11-16 02:45:17.995921','2022-11-16 02:45:17.995921','2655832e-f009-4b73-a535-64c3a22e558f','d796884d-a8c4-4525-81e7-54a3b6099eac'),
('4bb1ea40-c002-4fd9-8f60-82b3aa9459d7',NULL,'2022-11-16 02:45:17.997153','2022-11-16 02:45:17.997153','ecea9d15-05ba-4a4e-9787-54210e3b98ce','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('8fb91cd7-f461-49fb-8b7b-9ef9b524789d',NULL,'2022-11-16 02:45:17.998377','2022-11-16 02:45:17.998377','ecea9d15-05ba-4a4e-9787-54210e3b98ce','01fb6322-d727-47e3-a242-5039ea4732fc'),
('3e381f98-8303-4274-ba85-dd3d91b03178',NULL,'2022-11-17 17:36:17.856267','2022-11-17 17:36:17.856267','ecea9d15-05ba-4a4e-9787-54210e3b98ce','11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('f5b780db-0ac9-4518-9856-ed4d655c4080',NULL,'2022-11-17 17:36:17.856267','2022-11-17 17:36:17.856267','8d7c99c0-3253-4286-93a9-e7554cb327ef','11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('60317315-4e43-4876-9763-ac11358fe3be',NULL,'2022-11-17 17:36:17.856267','2022-11-17 17:36:17.856267','eeae589a-94e3-49ac-a94c-fcfb084152b2','11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('afa89e38-6f39-4317-83de-2eb7fc6b1a4b',NULL,'2022-11-17 17:36:17.856267','2022-11-17 17:36:17.856267','608ea227-45a5-4c8a-af43-6c7280d96340','11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('204f81e3-f504-4590-bbac-d18a0367a300',NULL,'2022-11-17 17:36:17.856267','2022-11-17 17:36:17.856267','3f705527-0455-4e67-a585-6c1f23726fff','11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('7f56677b-ce3a-4eb7-821a-7e76e8d8c1c3',NULL,'2022-11-17 17:36:30.869365','2022-11-17 17:36:30.869365','ecea9d15-05ba-4a4e-9787-54210e3b98ce','d796884d-a8c4-4525-81e7-54a3b6099eac'),
('b664cc94-e919-44ca-9486-0779f4f485a2',NULL,'2022-11-17 17:36:30.869365','2022-11-17 17:36:30.869365','8d7c99c0-3253-4286-93a9-e7554cb327ef','d796884d-a8c4-4525-81e7-54a3b6099eac'),
('efe6202f-83e5-4c6a-b33c-2633269bb6d0',NULL,'2022-11-17 17:36:30.869365','2022-11-17 17:36:30.869365','eeae589a-94e3-49ac-a94c-fcfb084152b2','d796884d-a8c4-4525-81e7-54a3b6099eac'),
('8480b031-0981-43ee-ae9b-7a553b0decae',NULL,'2022-11-17 17:36:30.869365','2022-11-17 17:36:30.869365','608ea227-45a5-4c8a-af43-6c7280d96340','d796884d-a8c4-4525-81e7-54a3b6099eac'),
('9398494e-1d50-4cec-9b30-c5245d2e89d1',NULL,'2022-11-17 17:36:30.869365','2022-11-17 17:36:30.869365','3f705527-0455-4e67-a585-6c1f23726fff','d796884d-a8c4-4525-81e7-54a3b6099eac'),
('9742e6c2-03ca-4956-b93e-31dc5e829bc3',NULL,'2022-11-17 17:36:44.72235','2022-11-17 17:36:44.72235','2655832e-f009-4b73-a535-64c3a22e558f','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('2528efdb-dab7-4a8b-840c-2512222a77a4',NULL,'2022-11-17 17:36:44.72235','2022-11-17 17:36:44.72235','8d7c99c0-3253-4286-93a9-e7554cb327ef','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('efbf9aea-c855-40b0-8585-60430320c2a9',NULL,'2022-11-17 17:36:44.72235','2022-11-17 17:36:44.72235','eeae589a-94e3-49ac-a94c-fcfb084152b2','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('6c81fd09-671a-4954-a552-cbedacc2a57f',NULL,'2022-11-17 17:36:44.72235','2022-11-17 17:36:44.72235','608ea227-45a5-4c8a-af43-6c7280d96340','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('d41e060c-cc12-42a6-93ae-1ea0ca93e581',NULL,'2022-11-17 17:36:44.72235','2022-11-17 17:36:44.72235','3f705527-0455-4e67-a585-6c1f23726fff','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('c8e4b92c-7800-4fc7-b4f7-38b2916a84fa',NULL,'2022-11-17 17:36:55.66746','2022-11-17 17:36:55.66746','d1b1dcbe-f4ed-49cc-a16b-56c9a3145c14','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('600cf853-10bc-4e31-a570-53441c10df4c',NULL,'2022-11-17 17:36:55.66746','2022-11-17 17:36:55.66746','7bcf4b45-0279-4da2-84e4-eec5d9d05ba3','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('06a32e7f-adf7-4b54-82e3-bde16ea42f45',NULL,'2022-11-17 17:36:55.66746','2022-11-17 17:36:55.66746','37e00d47-de69-4ac8-991b-cf3e39565c00','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('f14c40d8-3b0b-4dc0-a7cf-4113c261a753',NULL,'2022-11-17 17:36:55.66746','2022-11-17 17:36:55.66746','dcfee24b-306d-4609-9c24-a4021639a11b','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('6fed5faf-4469-4000-9460-0309b66e12bb',NULL,'2022-11-17 17:36:55.66746','2022-11-17 17:36:55.66746','a1f25e3e-cf40-4d97-a4d5-c27ee566b8c5','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('3ae11671-c1ba-4fcd-bcca-39fa5acf743c',NULL,'2022-11-17 17:36:57.353623','2022-11-17 17:36:57.353623','d389d35e-c548-42cf-9f29-2a599969a8f2','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('039ac595-6be8-4e89-bb9b-d98b37b91477',NULL,'2022-11-17 17:36:57.353623','2022-11-17 17:36:57.353623','498d5805-dd64-4643-902d-95067d6e5ab5','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('63cbf470-4ed3-464f-bbc1-42da8b83f497',NULL,'2022-11-17 17:36:57.353623','2022-11-17 17:36:57.353623','fae443f8-e8ff-4eec-b86c-98216d7662d8','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('ce54e86f-e49d-488a-86f8-80b85709d91d',NULL,'2022-11-17 17:36:57.353623','2022-11-17 17:36:57.353623','ac318f73-0c8e-43ed-b00e-d230670dc400','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('c47438ca-86b2-4026-997c-a40a57ed3d7f',NULL,'2022-11-17 17:36:57.353623','2022-11-17 17:36:57.353623','7fbfb59b-e00e-46b2-85cd-64f9f9942daa','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('def814fb-ff7c-4663-bb99-a359c3c0b24e',NULL,'2022-11-17 17:37:04.691623','2022-11-17 17:37:04.691623','2655832e-f009-4b73-a535-64c3a22e558f','01fb6322-d727-47e3-a242-5039ea4732fc'),
('e73928ba-cd5a-4edd-8ea7-18859d6dd1c9',NULL,'2022-11-17 17:37:04.691623','2022-11-17 17:37:04.691623','8d7c99c0-3253-4286-93a9-e7554cb327ef','01fb6322-d727-47e3-a242-5039ea4732fc'),
('17512e68-dafe-481c-aedc-f7f9ec0b3294',NULL,'2022-11-17 17:37:04.691623','2022-11-17 17:37:04.691623','eeae589a-94e3-49ac-a94c-fcfb084152b2','01fb6322-d727-47e3-a242-5039ea4732fc'),
('008619ad-89cd-4da3-841c-c0eb7f890d9f',NULL,'2022-11-17 17:37:04.691623','2022-11-17 17:37:04.691623','608ea227-45a5-4c8a-af43-6c7280d96340','01fb6322-d727-47e3-a242-5039ea4732fc'),
('cfc6049e-6195-4199-8b89-3280558e03f2',NULL,'2022-11-17 17:37:04.691623','2022-11-17 17:37:04.691623','3f705527-0455-4e67-a585-6c1f23726fff','01fb6322-d727-47e3-a242-5039ea4732fc');

INSERT INTO public.project_slot_positions (id,deleted_at,created_at,updated_at,project_slot_id,position_id) VALUES
('44cb1c10-1a39-4c92-bb85-532b7456b456',NULL,'2022-11-18 01:11:32.830016','2022-11-18 01:11:32.830016','f32d08ca-8863-4ab3-8c84-a11849451eb7','11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('28afce2a-e01b-4230-8977-429a3d04e47b',NULL,'2022-11-18 01:11:41.429707','2022-11-18 01:11:41.429707','bdc64b18-4c5f-4025-827a-f5b91d599dc7','d796884d-a8c4-4525-81e7-54a3b6099eac'),
('7fc83d8e-ce33-470e-9ef7-c8249c20b2ed',NULL,'2022-11-18 01:11:42.646576','2022-11-18 01:11:42.646576','1406fcce-6f90-4e0f-bea1-c373e2b2b5b1','d796884d-a8c4-4525-81e7-54a3b6099eac'),
('f2de0d25-8f8b-433d-af8c-2a2df454b2e7',NULL,'2022-11-18 01:11:52.215061','2022-11-18 01:11:52.215061','b25bd3fa-eb6d-49d5-b278-7aacf4594f79','dac16ce6-9e5a-4ff3-9ea2-fdea4853925e'),
('e043f0cf-4a25-40c3-a6ee-e68029387dd0',NULL,'2022-11-18 01:12:05.162737','2022-11-18 01:12:05.162737','ce379dc0-95be-471a-9227-8e045a5630af','11ccffea-2cc9-4e98-9bef-3464dfe4dec8');

INSERT INTO public.project_member_positions (id,deleted_at,created_at,updated_at,project_member_id,position_id) VALUES
('df34612a-df98-4618-8e29-c56916347302',NULL,'2022-11-18 01:14:36.711168','2022-11-18 01:14:36.711168','cb889a9c-b20c-47ee-83b8-44b6d1721aca','11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('fe6dba87-98b3-466a-86f1-2c98b62b83e1',NULL,'2022-11-18 01:14:37.611091','2022-11-18 01:14:37.611091','7310b51a-3855-498b-99ab-41aa82934269','11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('34aada0f-69ec-420e-a4e6-2db8d9c6d8e9',NULL,'2022-11-18 01:14:38.042046','2022-11-18 01:14:38.042046','35149aab-0506-4eb3-9300-c706ccbf2bde','11ccffea-2cc9-4e98-9bef-3464dfe4dec8'),
('3a35f08a-5af8-4145-bbb9-3eea86ee7783',NULL,'2022-11-18 01:14:47.124502','2022-11-18 01:14:47.124502','5a9a07aa-e8f3-4b62-b9ad-0f057866dc6c','d796884d-a8c4-4525-81e7-54a3b6099eac'),
('913e5ecb-0053-4ff3-aa23-64dd96f216a6',NULL,'2022-11-18 01:14:47.556013','2022-11-18 01:14:47.556013','fcd5c16f-40fd-48b6-9649-410691373eea','d796884d-a8c4-4525-81e7-54a3b6099eac');

INSERT INTO public.currencies (id, deleted_at, created_at, updated_at, name, symbol, locale, type) VALUES
('06a699ed-618b-400b-ac8c-8739956fa8e7', NULL, '2019-02-20 04:24:14.967209+00', '2019-02-20 04:24:14.967209+00', 'GBP', '£', 'en-gb', 'fiat'),
('0a6f4a2e-a097-4f7e-ae65-bfee3298e5cc', NULL, '2020-11-23 08:59:49.802559+00', '2020-11-23 08:59:49.802559+00', 'TEN', NULL, NULL, 'crypto'),
('1c7dcbe2-6984-461d-8ed9-537676f2b590', NULL, '2021-07-21 18:10:24.194492+00', '2021-07-21 18:10:24.194492+00', 'USD', '$', 'en-us', 'crypto'),
('283e559f-a7e7-4aa7-9e08-4806d6c07016', NULL, '2019-02-20 04:24:14.967209+00', '2019-02-20 04:24:14.967209+00', 'EUR', '€', 'en-gb', 'fiat'),
('2de81125-a947-4fea-a006-2c60e7ec01ed', NULL, '2020-02-28 07:18:02.349700+00', '2020-02-28 07:18:02.349700+00', 'CAD', 'c$', 'en-ca', 'fiat'),
('7037bdb6-584e-4e35-996d-ef28a243f48a', NULL, '2019-02-11 06:14:51.305496+00', '2019-02-11 06:14:51.305496+00', 'VND', 'đ', 'vi-vn', 'fiat'),
('bf256e69-28b0-4d9f-bf48-3662854157a9', NULL, '2019-10-28 14:14:52.302051+00', '2019-10-28 14:14:52.302051+00', 'SGD', 's$', 'en-sg', 'fiat'),
('f00498e4-7a4c-4f61-b126-b84b5faeee06', NULL, '2019-02-11 06:14:51.305496+00', '2019-02-11 06:14:51.305496+00', 'USD', '$', 'en-us', 'fiat');

INSERT INTO public.work_units (id, deleted_at, created_at, updated_at, name, status, type, source_url, project_id, source_metadata) VALUES 
('4797347d-21e0-4dac-a6c7-c98bf2d6b27c', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'Fortress API', 'active', 'development', 'https://github.com/dwarvesf/fortress-api', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '[]'),
('69b32f7e-0433-4566-a801-72909172940e', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'Fortress Web', 'archived', 'management', 'https://github.com/dwarvesf/fortress-web', 'dfa182fc-1d2d-49f6-a877-c01da9ce4207', '[]');

INSERT INTO public.work_unit_stacks (id, deleted_at, created_at, updated_at, stack_id, work_unit_id) VALUES 
('95851b98-c8d0-46f6-b6b0-8dd2037f44d6', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', '0ecf47c8-cca4-4c30-94bb-054b1124c44f', '4797347d-21e0-4dac-a6c7-c98bf2d6b27c'),
('f1ddeeb2-ad44-4c97-a934-86ad8f24ca57', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'fa0f4e46-7eab-4e5c-9d31-30489e69fe2e', '4797347d-21e0-4dac-a6c7-c98bf2d6b27c'),
('9b7c9d01-75a3-4386-93a7-4ff099887847', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'b403ef95-4269-4830-bbb6-8e56e5ec0af4', '69b32f7e-0433-4566-a801-72909172940e'),
('b851f3bc-a758-4e28-834b-0d2c0a04bf71', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'fa0f4e46-7eab-4e5c-9d31-30489e69fe2e', '69b32f7e-0433-4566-a801-72909172940e');

INSERT INTO public.work_unit_members (id, deleted_at, created_at, updated_at, joined_date, left_date, status, project_id, employee_id, work_unit_id) VALUES 
('303fd2e5-0b4d-401c-b5fa-74820991e6c0', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', '2022-11-29', NULL, 'active', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '2655832e-f009-4b73-a535-64c3a22e558f', '4797347d-21e0-4dac-a6c7-c98bf2d6b27c'),
('f79ae054-4ab4-41cd-aa5f-c871887cc35c', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', '2022-11-29', NULL, 'active', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', '69b32f7e-0433-4566-a801-72909172940e'),
('7e4da4ac-241f-4af8-b0a0-f59e5a64065b', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', '2022-11-29', NULL, 'active', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '8d7c99c0-3253-4286-93a9-e7554cb327ef', '69b32f7e-0433-4566-a801-72909172940e'),
('93954bcd-d5e9-4c4c-ad30-7de5fd332a80', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', '2022-11-29', NULL, 'inactive', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'eeae589a-94e3-49ac-a94c-fcfb084152b2', '69b32f7e-0433-4566-a801-72909172940e'),
('e14f68f8-7ed5-4a59-9df8-275573537861', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', '2022-11-29', NULL, 'inactive', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '608ea227-45a5-4c8a-af43-6c7280d96340', '4797347d-21e0-4dac-a6c7-c98bf2d6b27c'),
('799708d7-855b-4a42-b169-7c9891f0b218', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', '2022-11-29', NULL, 'inactive', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', '4797347d-21e0-4dac-a6c7-c98bf2d6b27c');

INSERT INTO public.feedback_events (id, deleted_at, created_at, updated_at, type, subtype, status, created_by, start_date, end_date) VALUES 
('9b3480be-86a2-4ff9-84d8-545a4146122b', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'feedback', 'comment', NULL, '2655832e-f009-4b73-a535-64c3a22e558f', NULL, NULL),
('8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'survey', 'peer-review', NULL, '2655832e-f009-4b73-a535-64c3a22e558f', '2022-11-29 08:03:33.233262', '2023-05-29 08:03:33.233262'),
('d97ee823-f7d5-418b-b281-711cb1d8e947', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'survey', 'work', NULL, '2655832e-f009-4b73-a535-64c3a22e558f', NULL, NULL),
('53546ea4-1d9d-4216-96b2-75f84ec6d750', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'survey', 'engagement', NULL, 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', NULL, NULL),
('163fdda2-2dce-4618-9849-7c8475dcc9c1', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'feedback', 'appreciation', NULL, 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', NULL, NULL);

INSERT INTO public.employee_event_topics (id, deleted_at, created_at, updated_at, title, event_id, employee_id, project_id) VALUES
('e4a33adc-2495-43cf-b816-32feb8d5250d', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'Review Nguyen', '8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2', '38a00d4a-bc45-41de-965a-adc674ab82c9', NULL),
('9cf93fc1-5a38-4e2a-87de-41634b65fc87', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'Self review Fortress by Thanh', 'd97ee823-f7d5-418b-b281-711cb1d8e947', '2655832e-f009-4b73-a535-64c3a22e558f', '8dc3be2e-19a4-4942-8a79-56db391a0b15'),
('11121775-118f-4896-8246-d88023b22c7a', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'Engagement Review by Huy', '53546ea4-1d9d-4216-96b2-75f84ec6d750', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', NULL),
('a3b2cd0b-b327-4118-83f2-29350e678379', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'Huy want to give Nguyen shoutout', '163fdda2-2dce-4618-9849-7c8475dcc9c1', '38a00d4a-bc45-41de-965a-adc674ab82c9', NULL);

INSERT INTO public.employee_event_reviewers (id, deleted_at, created_at, updated_at, employee_event_topic_id, reviewer_id, status, relationship, is_shared, is_read, event_id) VALUES 
('bc9a5715-9723-4a2f-ad42-0d0f19a80b4d', NULL, '2022-12-05 16:26:15.411511', '2022-12-05 16:26:15.411511', 'e4a33adc-2495-43cf-b816-32feb8d5250d', '2655832e-f009-4b73-a535-64c3a22e558f', 'done', 'peer', false, false, '8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2'),
('e96999f5-b3f9-420c-9d9f-aa64e3e889bf', NULL, '2022-12-05 16:26:15.411511', '2022-12-05 16:26:15.411511', 'e4a33adc-2495-43cf-b816-32feb8d5250d', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', 'draft', 'line-manager', false, false, '8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2'),
('c994db17-384a-4331-8944-b8ac0070ac3f', NULL, '2022-12-05 16:26:15.411511', '2022-12-05 16:26:15.411511', 'e4a33adc-2495-43cf-b816-32feb8d5250d', '37e00d47-de69-4ac8-991b-cf3e39565c00', 'draft', 'peer', false, false, '8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2'),
('789f1163-f157-4df3-9764-8100277cacba', NULL, '2022-12-05 16:26:15.411511', '2022-12-05 16:26:15.411511', '9cf93fc1-5a38-4e2a-87de-41634b65fc87', '2655832e-f009-4b73-a535-64c3a22e558f', 'draft', 'self', false, false, 'd97ee823-f7d5-418b-b281-711cb1d8e947'),
('41f6a7fb-baa2-4e61-8035-e36d752dc611', NULL, '2022-12-05 16:26:15.411511', '2022-12-05 16:26:15.411511', '11121775-118f-4896-8246-d88023b22c7a', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', 'draft', 'self', true, true, '53546ea4-1d9d-4216-96b2-75f84ec6d750'),
('1a5eebbb-f3f7-40a7-9c95-2240df3aecef', NULL, '2022-12-05 16:26:15.411511', '2022-12-05 16:26:15.411511', 'a3b2cd0b-b327-4118-83f2-29350e678379', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', 'draft', 'peer', false, false, '163fdda2-2dce-4618-9849-7c8475dcc9c1');

INSERT INTO public.employee_event_questions (id, deleted_at, created_at, updated_at, employee_event_reviewer_id, content, answer, note, question_id, type, "order", event_id) VALUES 
('4adf1a24-f89e-4286-aeab-090bf5e9a030', NULL, '2022-12-05 16:33:28.085352', '2022-12-05 16:33:28.085352', 'bc9a5715-9723-4a2f-ad42-0d0f19a80b4d', 'Does this employee effectively communicate with othes?', 'good', NULL, 'da5dbdd5-8e1e-4ae7-8bb8-ab007f2580aa', 'general', 1, '8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2'),
('805b3bdb-bb90-44eb-a1ed-1ddf8bda8bd9', NULL, '2022-12-05 16:33:28.085352', '2022-12-05 16:33:28.085352', 'bc9a5715-9723-4a2f-ad42-0d0f19a80b4d', 'How effective of a leader is this person, either through direct management or influence?', 'good', NULL, '7d95e035-81d6-49d7-bed4-3a83bf2e34d6', 'general', 2, '8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2'),
('99862c14-a9eb-40d1-8e09-0d02ee8d0b67', NULL, '2022-12-05 16:33:28.085352', '2022-12-05 16:33:28.085352', 'bc9a5715-9723-4a2f-ad42-0d0f19a80b4d', 'Does this person find creative solutions, and own the solution to problems? Are they proactive or reactive?', 'good', NULL, 'd36e84c5-d7a4-4d5f-ada1-f6b9ddb58f51', 'general', 3, '8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2');

INSERT INTO public.questions (id, deleted_at, created_at, updated_at, category, subcategory, content, type, "order") VALUES 
('da5dbdd5-8e1e-4ae7-8bb8-ab007f2580aa', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'feedback', 'peer-review', 'Does this employee effectively communicate with others?', 'general', 1),
('7d95e035-81d6-49d7-bed4-3a83bf2e34d6', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'feedback', 'peer-review', 'How effective of a leader is this person', 'general', 2),
('d36e84c5-d7a4-4d5f-ada1-f6b9ddb58f51', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'feedback', 'peer-review', 'Does this person find creative solutions', 'general', 3),
('f03432ba-c024-467e-8059-a5bb2b7f783d', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'feedback', 'peer-review', 'How would you rate the quality of the employee''s work?', 'general', 4),
('d2bb48c1-e8d6-4946-a372-8499907b7328', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'feedback', 'peer-review', 'How well does this person set and meet deadlines?', 'general', 5),
('be86ce52-803b-403f-b059-1a69492fe3d4', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'feedback', 'peer-review', 'How well does this person embody our culture?', 'general', 6),
('51eab8c7-61ba-4c56-be39-b72eb6b89a52', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'feedback', 'peer-review', 'If you could give this person one piece of constructive advice to make them more effective in their role', 'general', 7),
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
