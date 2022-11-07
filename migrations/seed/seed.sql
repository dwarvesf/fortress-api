INSERT INTO "public"."employees" ("id", "deleted_at", "created_at", "updated_at", "full_name", "display_name", "gender", "team_email", "personal_email", "avatar", "phone_number", "address", "mbti", "horoscope", "passport_photo_front", "passport_photo_back", "identity_card_photo_front", "identity_card_photo_back", "date_of_birth", "working_status", "joined_date", "left_date", "basecamp_id", "basecamp_attachable_sgid", "gitlab_id", "github_id", "discord_id", "wise_recipient_email", "wise_recipient_name", "wise_recipient_id", "wise_account_number", "wise_currency", "local_bank_branch", "local_bank_number", "local_bank_currency", "local_branch_name", "local_bank_recipient_name") VALUES
('2655832e-f009-4b73-a535-64c3a22e558f', NULL, '2022-11-02 09:52:34.586566', '2022-11-02 09:52:34.586566', 'Phạm Đức Thành', 'Thanh Pham', 'Male', 'thanh@d.foundation', 'thanhpham@gm', 'https://s3-ap-southeast-1.amazonaws.com/fortress-images/5153574695663955944.png', '0985586998', '102/61 Phan Huy Ich, Tan Binh District, Ho Chi Minh, Vietnam', 'INTJ-A', 'Libra', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
('ecea9d15-05ba-4a4e-9787-54210e3b98ce', NULL, '2022-11-02 09:50:55.320669', '2022-11-02 09:50:55.320669', 'Nguyễn Hoàng Huy', 'Huy Nguyen', 'Male', 'huy@d.foundation', 'hoanghuy2908@gmail.com', 'https://s3-ap-southeast-1.amazonaws.com/fortress-images/2830497479497502617.png', '0383830083', 'chung cư Sunview Town, Gò Dưa, Thủ Đức', 'Defender', 'Virgo', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL);

INSERT INTO public.roles (id, deleted_at, created_at, updated_at, name, code) VALUES
    ('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Admin', 'admin'),
    ('d796884d-a8c4-4525-81e7-54a3b6099eac', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Member', 'member');

INSERT INTO public.positions (id, deleted_at, created_at, updated_at, name, code) VALUES 
    ('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Frontend', 'frontend'),
    ('d796884d-a8c4-4525-81e7-54a3b6099eac', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Backend', 'backend'),
    ('dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Devops', 'devops'),
    ('01fb6322-d727-47e3-a242-5039ea4732fc', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Blockchain', 'blockchain'),
    ('39735742-829b-47f3-8f9d-daf0983914e5', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Project-Management', 'project-management');

INSERT INTO public.account_statuses (id, deleted_at, created_at, updated_at, name, code) VALUES
    ('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Onboarding', 'onboarding'),
    ('d796884d-a8c4-4525-81e7-54a3b6099eac', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Probation', 'probation'),
    ('dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Active', 'active'),
    ('01fb6322-d727-47e3-a242-5039ea4732fc', null, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'On Leave', 'on-leave');

INSERT INTO "public"."countries" ("id", "deleted_at", "created_at", "updated_at", "name", "code", "cities") VALUES 
('4ef64490-c906-4192-a7f9-d2221dadfe4c',NULL,'2022-11-08 08:06:56.068148','2022-11-08 08:06:56.068148','Vietnam','+84','["Hồ Chí Minh", "An Giang", "Bà Rịa - Vũng Tàu", "Bình Dương", "Bình Định", "Bình Phước", "Bình Thuận", "Bạc Liêu", "Bắc Giang", "Bắc Kạn", "Bắc Ninh", "Bến Tre", "Cao Bằng", "Cà Mau", "Cần Thơ", "Điện Biên", "Đà Nẵng", "Đắk Lắk", "Đồng Nai", "Đắk Nông", "Đồng Tháp", "Gia Lai", "Hoà Bình", "Hà Giang", "Hà Nam", "Hà Nội", "Hà Tĩnh", "Hải Dương", "Hải Phòng", "Hậu Giang", "Hưng Yên", "Khánh Hòa", "Kiên Giang", "Kon Tum", "Lai Châu", "Lâm Đồng", "Lạng Sơn", "Lào Cai", "Long An", "Nam Định", "Nghệ An", "Ninh Bình", "Ninh Thuận", "Phú Thọ", "Phú Yên", "Quảng Bình", "Quảng Nam", "Quảng Ngãi", "Quảng Ninh", "Quảng Trị", "Sóc Trăng", "Sơn La", "Thanh Hóa", "Thái Bình", "Thái Nguyên", "Thừa Thiên Huế", "Tiền Giang", "Trà Vinh", "Tuyên Quang", "Tây Ninh", "Vĩnh Long", "Vĩnh Phúc", "Yên Bái"]'),
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
