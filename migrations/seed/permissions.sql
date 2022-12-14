INSERT INTO public.permissions (id, deleted_at, created_at, updated_at, name, code) VALUES
('495c96ae-60f9-4c57-bc96-9504d0fedde6', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Auth Read', 'auth.read'), --
('01ed1076-4028-4fdf-9a92-cb57a8e041af', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read', 'employees.read'), --
('dc6fde9f-0b49-46d6-96bb-93be669b502b', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read Full Personal Info', 'employees.read.personalInfo.fullRead'), --
('f75677d6-3e22-45d4-b921-81d6a3645157', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read Full General Info', 'employees.read.generalInfo.fullRead'), --
('a7549f30-987c-47d1-9266-452f3cfc68b7', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Create', 'employees.create'),
('e4e68398-da67-438f-9fb0-07a779b504a0', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Edit', 'employees.edit'),
('33279d12-6daa-41d3-a037-9f805e8ebf61', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Delete', 'employees.delete'),
('834ce06a-2797-4974-bbdd-2dfffc431c5f', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read Filter By All Statuses', 'employees.read.filterByAllStatuses'),
('09a5fa4c-ad07-4dec-a16d-34e0f567ef1d', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Create', 'projects.create'),
('35bb795a-0b9f-428c-861d-48c1e8d4e73a', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Read', 'projects.read'), --
('80a2c800-02ea-4264-9289-57b92e911097', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Edit', 'projects.edit'),
('6797aeac-7d5e-4216-b260-a8a3f62d3cf6', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Member Create', 'projectMembers.create'),
('57ead3c5-b09f-4cd8-ac38-d9c0d4654af5', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Member Read', 'projectMembers.read'), --
('58719469-e01d-4667-bafa-12020931e317', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Member Edit', 'projectMembers.edit'),
('1a1303f3-a929-41b3-be7f-d372e6aab87a', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Member Delete', 'projectMembers.delete'),
('819adaec-f4a2-4438-937d-fcd9cc15d76b', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Work Unit Create', 'projectWorkUnits.create'),
('7c2d5fef-3096-4b68-bd76-f26cb2a3baea', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Work Unit Read', 'projectWorkUnits.read'), --
('cb90c56a-60a8-4345-8001-bb7ab172b302', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Work Unit Edit', 'projectWorkUnits.edit'),
('6eb3bf29-87dc-490b-85b4-e591d416ac8f', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Work Unit Delete', 'projectWorkUnits.delete'),
('615cf3a6-89cb-4200-81b4-47542ae8b145', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Feedback Create', 'feedbacks.create'), --
('cc88051f-7eea-468e-a9ac-91757dde2581', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Feedback Read', 'feedbacks.read'), --
('3b542878-f597-421f-b4ac-5f71b0e22ddb', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Feedback Edit', 'feedbacks.edit'),
('2969fd2a-ae44-414a-85ac-7a930c6de987', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Feedback Delete', 'feedbacks.delete'),
('37219e6a-34f1-4d68-8727-db90ddd6f97e', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Event Question Create', 'employeeEventQuestions.create'),
('bfd6144d-0161-41f2-9cf5-3467d2a505fd', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Event Question Edit', 'employeeEventQuestions.edit'),
('be88150b-3ad5-4cef-8716-90873d194207', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Event Question Read', 'employeeEventQuestions.read'),
('23fe9fa0-d848-491f-8b6d-04ada45b9c51', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Event Question Delete', 'employeeEventQuestions.delete'),
('d5c1142a-86a6-4347-bdef-2ad41492b738', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Survey Create', 'surveys.create'),
('ef32d604-42ec-40f8-9c34-08431d2c20d8', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Survey Read', 'surveys.read'),
('f5906795-60d1-4e5f-88f2-e15f2fae2327', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Survey Edit', 'surveys.edit'),
('dfc32d81-abfa-49ce-a8f6-cdc83c8da78b', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Survey Delete', 'surveys.delete'),
('ed107409-30c5-4555-ba2e-7d77ee8027dc', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Mentees Create', 'employeeMentees.create'),
('0b8bc573-c8a8-49f4-8274-486990e76540', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Mentees Read', 'employeeMentees.read'), --
('910cd3bd-3b9b-4243-899f-86ed47e9d1c9', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Mentees Edit', 'employeeMentees.edit'),
('967f2d24-1bd9-4b62-8136-7fed60dee6d9', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Mentees Delete', 'employeeMentees.delete'),
('050ee781-7f32-4d79-b3d3-1e9d550d4a8f', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Metadata Create', 'metadata.create'),
('476156c1-70d8-4eb1-adc5-f8f5ec209666', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Metadata Read', 'metadata.read'),
('a87e99c2-c40e-42e0-9310-943f2faa654b', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Metadata Edit', 'metadata.edit'),
('befac4bf-bb20-4b46-9e56-f8175c93b191', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Metadata Delete', 'metadata.delete'),
('f632f18c-98da-4c1a-b2c6-9ccd0c7c5a39', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Roles Create', 'employeeRoles.create'),
('5c5919b5-e74b-4854-bdf6-b91396a8317c', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Roles Read', 'employeeRoles.read'),
('68fe5bcf-38ae-446a-8cbb-8f88b5c6eb44', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Roles Edit', 'employeeRoles.edit'),
('98744d43-f14b-4235-8dfc-5265668daa21', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Roles Delete', 'employeeRoles.delete');
