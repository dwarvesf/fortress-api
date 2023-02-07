-- +migrate Up
-- +migrate StatementBegin
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = 'd13d9491-6d29-46fa-bca7-c14938978713';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '656dd865-39c1-4ed8-b72e-63d0e89c3679';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '656dd866-39c1-4ed8-b72e-63d0e89c3679';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '5a8f3d89-04a0-4a0e-bdc8-73045ced6a08';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '0b2068a7-ffb3-4feb-a8f4-ca3a80d41290';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = 'e0d8c920-f4f1-4309-aabf-1d83f792f45b';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '656dd867-39c1-4ed8-b72e-63d0e89c3679';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = 'e84c1860-6991-44a1-b687-fec078b842eb';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = 'b22f3f48-2b45-4855-9308-85588c15a6de';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '06126b41-b20f-4ca2-933a-06fdb01fe362';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = 'b7078b45-bc29-44cb-b3bb-cef86cb36926';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = 'ee219b4e-e4dc-4782-b590-03f799cd41ab';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '8d6a772a-0544-4b59-a279-d97d0e855966';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '9d70a553-2d5d-4a01-a7ed-ff898af82f1a';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = 'ec086e4a-2167-4924-adfd-84be02edebe9';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '69bf5adf-7ba2-4abc-b87e-9a68668a267e';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '2641297b-c632-4c22-9a42-61291d621552';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '3c420751-bb9a-4878-896e-2f10f3a633d6';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '40f33d74-91ae-4eec-ba35-8d330376d6e1';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '40a5e3c7-22b6-4954-be91-89c712143a86';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '63d163a7-e9f5-4210-a685-151061fe9c29';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = 'a8f34385-61f8-46f0-9403-4b05e37cd8e3';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = 'ef4d0c00-97f9-4fc1-96c1-15a5c029b7c0';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = 'f68a4ab5-4893-445b-a214-8aafecfaf2c1';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '3e858c81-d661-4d4f-b913-e02dd6f4007e';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '97c15b2a-1a6d-406b-ae48-bfb1c4b09e59';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '95f9a38c-bf3b-40fb-9406-38a72e2aa556';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = 'aaf722a5-1ce5-4464-8c28-bcfd571e8bbc';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '9a91ff6b-7367-403f-b23d-4c16dabd6857';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '61e038b6-6b7b-444b-b9d8-7a976ca80c15';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = 'a5ddbb54-3faa-4f92-964a-3754928d3f21';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '2b20fcc7-fdda-4a1a-b620-f32115cc84c6';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = 'e4b4b0e2-7163-470c-95a9-90f45d3a62d4';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = '9e23a602-706e-4165-bfe7-355a8a1d456a';
DELETE FROM social_accounts WHERE type = 'notion' and employee_id = 'ce0eee19-b8fa-4c1b-bec4-7f4244c7eb3f';

INSERT INTO public.social_accounts (employee_id, type, account_id) VALUES 
('d13d9491-6d29-46fa-bca7-c14938978713', 'notion', 'bf91fcb9-cc64-40cc-a341-44a98a1429a5'),
('656dd865-39c1-4ed8-b72e-63d0e89c3679', 'notion', '73ca2dc1-07c1-4b71-8a82-afb02f8cbc14'),
('656dd866-39c1-4ed8-b72e-63d0e89c3679', 'notion', '3a366683-9851-4c4f-84af-ae9e390d70d6'),
('5a8f3d89-04a0-4a0e-bdc8-73045ced6a08', 'notion', '64dd7d0c-9fdb-4191-a5f1-2270ed440258'),
('0b2068a7-ffb3-4feb-a8f4-ca3a80d41290', 'notion', '87b8c080-f7bc-4d95-afa0-888b95518b99'),
('e0d8c920-f4f1-4309-aabf-1d83f792f45b', 'notion', '8ba5fbb3-6f99-4637-be43-f464891e5b88'),
('656dd867-39c1-4ed8-b72e-63d0e89c3679', 'notion', 'f6c8a51d-f2c5-4914-b1eb-a98d5e656eef'),
('e84c1860-6991-44a1-b687-fec078b842eb', 'notion', '2b643256-c18a-4eb5-912c-e29da3191820'),
('b22f3f48-2b45-4855-9308-85588c15a6de', 'notion', 'ed54d4c8-9d45-45e9-8cd2-e32b5c8f8844'),
('06126b41-b20f-4ca2-933a-06fdb01fe362', 'notion', 'a25db3d0-88f3-4814-aaf2-4876467788b3'),
('b7078b45-bc29-44cb-b3bb-cef86cb36926', 'notion', '42720e94-9f4e-4358-9b77-e69a4e09cf06'),
('ee219b4e-e4dc-4782-b590-03f799cd41ab', 'notion', 'c5e7f612-dd04-463f-b5e9-655e81099e41'),
('8d6a772a-0544-4b59-a279-d97d0e855966', 'notion', '3b1e16f0-72e4-4dc4-975d-cfa9e2fe6e77'),
('9d70a553-2d5d-4a01-a7ed-ff898af82f1a', 'notion', 'fc89fc4c-7a07-4b87-a2ba-1d7c5597150b'),
('ec086e4a-2167-4924-adfd-84be02edebe9', 'notion', '135043fc-3e2a-499b-86e7-8e32a7b700f7'),
('69bf5adf-7ba2-4abc-b87e-9a68668a267e', 'notion', '980ba262-2e76-4155-b95b-fbb1d7420326'),
('2641297b-c632-4c22-9a42-61291d621552', 'notion', 'ce689920-159c-4b12-81cc-47b00b52b625'),
('3c420751-bb9a-4878-896e-2f10f3a633d6', 'notion', '16591e20-1426-452e-b481-347387c54803'),
('40f33d74-91ae-4eec-ba35-8d330376d6e1', 'notion', '86ddbd3f-891b-4225-b620-2831feff7632'),
('40a5e3c7-22b6-4954-be91-89c712143a86', 'notion', 'bae0ffc4-4818-4a4e-aa01-32ed1bccb0d0'),
('63d163a7-e9f5-4210-a685-151061fe9c29', 'notion', '8fca1481-d287-436c-9776-2932d69345aa'),
('a8f34385-61f8-46f0-9403-4b05e37cd8e3', 'notion', 'edd3be21-1fa6-4a16-b054-c61eefc3eb56'),
('ef4d0c00-97f9-4fc1-96c1-15a5c029b7c0', 'notion', '195d94e2-0a92-40b6-aa0f-c1a9025da3e5'),
('f68a4ab5-4893-445b-a214-8aafecfaf2c1', 'notion', '29b84100-8070-4216-b2aa-cd8e0ce52cf9'),
('3e858c81-d661-4d4f-b913-e02dd6f4007e', 'notion', 'd50a8d8a-1c61-46e3-9532-491f692c6c53'),
('97c15b2a-1a6d-406b-ae48-bfb1c4b09e59', 'notion', '96384aa6-0abd-4526-baf0-9644c86332f3'),
('95f9a38c-bf3b-40fb-9406-38a72e2aa556', 'notion', 'e07ebee5-f6ea-4d57-bb83-586dbe087ef0'),
('aaf722a5-1ce5-4464-8c28-bcfd571e8bbc', 'notion', 'd3acd2a4-ef08-494f-871a-3bb4ba10659b'),
('9a91ff6b-7367-403f-b23d-4c16dabd6857', 'notion', 'e8d7994a-58e8-46c9-824a-b9e87e3ca310'),
('61e038b6-6b7b-444b-b9d8-7a976ca80c15', 'notion', '3545b8e8-d0d0-42d3-8923-88c577eb0ef9'),
('a5ddbb54-3faa-4f92-964a-3754928d3f21', 'notion', '8079cf90-7c4a-4278-915b-dc9a082dfb2f'),
('2b20fcc7-fdda-4a1a-b620-f32115cc84c6', 'notion', '3cbb184f-cec7-4c4f-a155-909e984e0700'),
('e4b4b0e2-7163-470c-95a9-90f45d3a62d4', 'notion', 'd68d8900-aa4e-4ffa-8212-babddb8c15c0'),
('9e23a602-706e-4165-bfe7-355a8a1d456a', 'notion', '5a3a0949-5eec-46d4-8407-d92774e2d987'),
('ce0eee19-b8fa-4c1b-bec4-7f4244c7eb3f', 'notion', '565407e6-6189-4199-aa81-491813c8c645');
-- +migrate StatementEnd
-- +migrate Down
SELECT TRUE;
