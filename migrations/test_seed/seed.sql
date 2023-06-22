INSERT INTO public.roles (id, deleted_at, created_at, updated_at, name, code, level) VALUES
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Admin', 'admin', 1),
('d796884d-a8c4-4525-81e7-54a3b6099eac', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Member', 'member', 2);

INSERT INTO public.permissions (id, deleted_at, created_at, updated_at, name, code) VALUES
('495c96ae-60f9-4c57-bc96-9504d0fedde6', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Auth Read', 'auth.read'),
('01ed1076-4028-4fdf-9a92-cb57a8e041af', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read', 'employees.read'),
('c353c4fd-5915-47f3-a59b-66ae75eae195', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read Get Active', 'employees.read.readActive'),
('20a442b2-9763-41d4-b3c9-e436b37bf534', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read', 'employees.fullAccess'),
('dc6fde9f-0b49-46d6-96bb-93be669b502b', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read Full Personal Info', 'employees.read.personalInfo.fullAccess'),
('f75677d6-3e22-45d4-b921-81d6a3645157', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read Full General Info', 'employees.read.generalInfo.fullAccess'),
('6add4087-e586-4313-b157-54d416bdc8d5', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read Full Project Belong to One Employee', 'employees.read.projects.fullAccess'),
('911df73b-b860-4319-8944-c274781591ca', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read Full Active Project Belong to One Employee', 'employees.read.projects.readActive'),
('5c1745b6-d920-47d2-986a-fe6c48802ace', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read Line Manager', 'employees.read.lineManager.fullAccess'),
('834ce06a-2797-4974-bbdd-2dfffc431c5f', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read Filter By All Statuses', 'employees.read.filterByAllStatuses'),
('11e16c63-87b2-4874-8961-c21716bdd97e', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Read Filter By All Statuses', 'employees.read.filterByProject'),
('a7549f30-987c-47d1-9266-452f3cfc68b7', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Create', 'employees.create'),
('e4e68398-da67-438f-9fb0-07a779b504a0', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Edit', 'employees.edit'),
('33279d12-6daa-41d3-a037-9f805e8ebf61', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Delete', 'employees.delete'),
('09a5fa4c-ad07-4dec-a16d-34e0f567ef1d', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Create', 'projects.create'),
('35bb795a-0b9f-428c-861d-48c1e8d4e73a', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Read', 'projects.read'),
('a8299847-4d10-41e9-8327-42e13e0672ae', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Read Full Access', 'projects.read.fullAccess'),
('80a2c800-02ea-4264-9289-57b92e911097', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Edit', 'projects.edit'),
('6118a1da-aa9a-40eb-b664-87a6969c5759', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Commission Rate Read', 'projects.commissionRate.read'),
('79b1a377-7b73-4bea-90c2-330bda11b635', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Commission Rate Edit', 'projects.commissionRate.edit'),
('6797aeac-7d5e-4216-b260-a8a3f62d3cf6', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Member Create', 'projectMembers.create'),
('57ead3c5-b09f-4cd8-ac38-d9c0d4654af5', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Member Read', 'projectMembers.read'),
('58719469-e01d-4667-bafa-12020931e317', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Member Edit', 'projectMembers.edit'),
('1a1303f3-a929-41b3-be7f-d372e6aab87a', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Member Delete', 'projectMembers.delete'),
('e33cc269-3a12-436e-aab7-75f14953388f', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Member Rate Read', 'projectMembers.rate.read'),
('2f6c957c-725b-4da4-ad6d-9a23a1e18a98', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Project Member Rate Edit', 'projectMembers.rate.edit'),
('819adaec-f4a2-4438-937d-fcd9cc15d76b', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Work Unit Create', 'projectWorkUnits.create'),
('7c7937f6-f3ae-47a4-8ee1-fb03f2a5197b', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Work Unit Create Full Access', 'projectWorkUnits.create.fullAccess'),
('7c2d5fef-3096-4b68-bd76-f26cb2a3baea', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Work Unit Read', 'projectWorkUnits.read'),
('d2d5dad6-efa3-4d20-9b30-34c6c933fa5e', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Work Unit Read Full Access', 'projectWorkUnits.read.fullAccess'),
('cb90c56a-60a8-4345-8001-bb7ab172b302', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Work Unit Edit', 'projectWorkUnits.edit'),
('8bae7a76-9f0e-4074-bfb9-31c3ad9c88e6', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Work Unit Edit Full Access', 'projectWorkUnits.edit.fullAccess'),
('6eb3bf29-87dc-490b-85b4-e591d416ac8f', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Work Unit Delete', 'projectWorkUnits.delete'),
('455368b3-abbe-4bc7-beb3-01f070addc14', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Work Unit Delete Full Access', 'projectWorkUnits.delete.fullAccess'),
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
('dfc32d81-abfa-49ce-a8f6-cdc83c8da78b', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Survey Delete', 'surveys.delete'),
('ed107409-30c5-4555-ba2e-7d77ee8027dc', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Mentees Create', 'employeeMentees.create'),
('0b8bc573-c8a8-49f4-8274-486990e76540', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Mentees Read', 'employeeMentees.read'),
('910cd3bd-3b9b-4243-899f-86ed47e9d1c9', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Mentees Edit', 'employeeMentees.edit'),
('967f2d24-1bd9-4b62-8136-7fed60dee6d9', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Mentees Delete', 'employeeMentees.delete'),
('050ee781-7f32-4d79-b3d3-1e9d550d4a8f', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Metadata Create', 'metadata.create'),
('476156c1-70d8-4eb1-adc5-f8f5ec209666', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Metadata Read', 'metadata.read'),
('a87e99c2-c40e-42e0-9310-943f2faa654b', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Metadata Edit', 'metadata.edit'),
('befac4bf-bb20-4b46-9e56-f8175c93b191', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Metadata Delete', 'metadata.delete'),
('f632f18c-98da-4c1a-b2c6-9ccd0c7c5a39', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Roles Create', 'employeeRoles.create'),
('5c5919b5-e74b-4854-bdf6-b91396a8317c', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Roles Read', 'employeeRoles.read'),
('68fe5bcf-38ae-446a-8cbb-8f88b5c6eb44', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Roles Edit', 'employeeRoles.edit'),
('98744d43-f14b-4235-8dfc-5265668daa21', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Employee Roles Delete', 'employeeRoles.delete'),
('4be48741-f8b2-4562-964e-65c081fec7a9', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Dashboard Create', 'dashboards.create'),
('f5854cfd-f990-46fd-a6cb-77b48a80b24b', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Dashboard Read', 'dashboards.read'),
('6f4fe3df-9ab8-46c1-a8b7-9458413980f3', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Dashboard Edit', 'dashboards.edit'),
('9c644140-c71d-4435-bd6b-a81171913046', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Invoice Create', 'invoices.create'),
('259fb434-5321-43be-b007-76f3c2dfbfcc', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Invoice Read', 'invoices.read'),
('fa6961f1-1cee-494f-8f1f-f4552b49b6fa', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Invoice Edit', 'invoices.edit'),
('60768a38-8ff1-493a-85ba-b62df98a60d4', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Invoice Delete', 'invoices.delete'),
('d93ede75-8fd5-4f88-b119-07493b757f4b', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Dashboard Delete', 'dashboards.delete'),
('a66215da-48dc-419d-8e2d-3f409aa038e3', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Dashboard Projects Read', 'dashboards.project.read'),
('36e9c7bd-c98f-4414-a8f4-164da5299874', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Dashboard Resources Read', 'dashboards.resources.read'),
('ca0af093-65ed-48d4-a6d1-62985a35fd8f', NULL, '2022-11-11 18:34:14.743263', '2022-11-11 18:34:14.743263', 'Dashboard Engagement Read', 'dashboards.engagement.read');

INSERT INTO public.role_permissions (id, deleted_at, created_at, updated_at, role_id, permission_id) VALUES
('6ef5b987-0a8d-42e8-b52c-fe683cd7a060', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '495c96ae-60f9-4c57-bc96-9504d0fedde6'),
('f73306cc-5b01-49bb-87af-a2c2af14bfae', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '01ed1076-4028-4fdf-9a92-cb57a8e041af'),
('64b381e6-039e-42dc-bc13-b5ed46bac99d', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'c353c4fd-5915-47f3-a59b-66ae75eae195'),
('9b63e271-261c-4f2f-9675-5b760fe6fa29', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '20a442b2-9763-41d4-b3c9-e436b37bf534'),
('642691b3-ac2e-45eb-9caa-0442e696d0da', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'dc6fde9f-0b49-46d6-96bb-93be669b502b'),
('bfed95b5-6dd7-4f67-91b3-10ef314f8d9c', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'f75677d6-3e22-45d4-b921-81d6a3645157'),
('db1f6ec9-11e5-427a-8b3d-1f93103c9648', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '6add4087-e586-4313-b157-54d416bdc8d5'),
('76ac9b8b-f8c4-488c-bc5b-6f7afc9e336a', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '911df73b-b860-4319-8944-c274781591ca'),
('2c373ed3-7464-4933-9778-70b34f893494', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '834ce06a-2797-4974-bbdd-2dfffc431c5f'),
('2de51d66-7ba8-4144-8db6-f55ad2b90966', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '11e16c63-87b2-4874-8961-c21716bdd97e'),
('43a5255c-f739-4de5-9014-e1531f88366c', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '5c1745b6-d920-47d2-986a-fe6c48802ace'),
('ef3e5ee2-69b5-44a6-85fd-0b8e5329ae6b', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'a7549f30-987c-47d1-9266-452f3cfc68b7'),
('6ad4a140-86ee-442b-8a7a-a0566dbab1f4', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'e4e68398-da67-438f-9fb0-07a779b504a0'),
('531e689e-9be2-442b-ac31-111f12e11a89', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '33279d12-6daa-41d3-a037-9f805e8ebf61'),
('29a63fdc-136d-4f5b-8103-32a16f3dc565', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '09a5fa4c-ad07-4dec-a16d-34e0f567ef1d'),
('b3c345fd-e800-4a64-9f8c-56158e5918c5', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '35bb795a-0b9f-428c-861d-48c1e8d4e73a'),
('4f023ef5-4b02-4254-a28b-a60d4e1e7efe', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'a8299847-4d10-41e9-8327-42e13e0672ae'),
('4a5fd07a-e308-4032-9a73-ace558e043c6', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '80a2c800-02ea-4264-9289-57b92e911097'),
('a7b14563-51a8-4df9-a976-830ee2b12d47', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '6797aeac-7d5e-4216-b260-a8a3f62d3cf6'),
('d08730ac-c3ff-44b9-983c-9effda8284ac', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '57ead3c5-b09f-4cd8-ac38-d9c0d4654af5'),
('e2558cc9-73cf-4f60-be90-2a39ee2e95d1', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '58719469-e01d-4667-bafa-12020931e317'),
('4d5ab598-e4e9-43ed-9b26-a4f42454c46d', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '1a1303f3-a929-41b3-be7f-d372e6aab87a'),
('e242544a-24da-4820-8d28-8aa67001033d', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '819adaec-f4a2-4438-937d-fcd9cc15d76b'),
('d33d31c3-217c-417e-9d42-8fc266540060', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '7c7937f6-f3ae-47a4-8ee1-fb03f2a5197b'),
('1aef6c68-cc6c-4db3-b78f-c40b714ca4aa', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '7c2d5fef-3096-4b68-bd76-f26cb2a3baea'),
('a5b34914-298b-4b29-912a-585478d39b44', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'd2d5dad6-efa3-4d20-9b30-34c6c933fa5e'),
('62256fae-d6af-4246-807f-83cd580b1ad3', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'cb90c56a-60a8-4345-8001-bb7ab172b302'),
('373882f2-3529-46f2-af14-8ca18611708d', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '8bae7a76-9f0e-4074-bfb9-31c3ad9c88e6'),
('05043830-fb1a-4558-9b5c-be9bea090cad', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '6eb3bf29-87dc-490b-85b4-e591d416ac8f'),
('063388c2-8da3-4812-b15b-9c8f4ad4f330', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '455368b3-abbe-4bc7-beb3-01f070addc14'),
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
('d1b42bd7-32f1-4f14-8dcc-009375ae196f', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'dfc32d81-abfa-49ce-a8f6-cdc83c8da78b'),
('067dbc36-ca1b-40be-bfc1-9289ea617a8d', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'ed107409-30c5-4555-ba2e-7d77ee8027dc'),
('be045620-f435-4220-850d-d4b2be887746', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '0b8bc573-c8a8-49f4-8274-486990e76540'),
('357c4d81-59bd-46d7-8c9b-78a7b697e52d', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '910cd3bd-3b9b-4243-899f-86ed47e9d1c9'),
('a6e38701-6efb-4a86-92b7-a6db7ada7064', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '967f2d24-1bd9-4b62-8136-7fed60dee6d9'),
('995e790d-b8bc-44ee-a04d-1cc7cae4e6c3', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '050ee781-7f32-4d79-b3d3-1e9d550d4a8f'),
('f77dd4e7-b472-4e90-baaa-35d79854f24f', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '476156c1-70d8-4eb1-adc5-f8f5ec209666'),
('080fd0f5-8e12-4b2d-936e-a012615bd563', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'a87e99c2-c40e-42e0-9310-943f2faa654b'),
('d052884a-f03f-4f27-a99b-8aff6edfe5f5', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'befac4bf-bb20-4b46-9e56-f8175c93b191'),
('77c22876-ba30-4446-b8d3-54f4d7d9fe9a', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'f632f18c-98da-4c1a-b2c6-9ccd0c7c5a39'),
('2125b270-507d-4760-8f08-84cff8b0794a', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '5c5919b5-e74b-4854-bdf6-b91396a8317c'),
('1e7e7b3c-fc0d-46ad-b822-4bce45d82ad1', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '68fe5bcf-38ae-446a-8cbb-8f88b5c6eb44'),
('8d13bab8-ca01-4ccd-b5ff-9c97cc969b0a', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '98744d43-f14b-4235-8dfc-5265668daa21'),
('86aae609-c38d-407c-9428-06e9cc812302', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '4be48741-f8b2-4562-964e-65c081fec7a9'),
('f8054472-8545-41ed-a654-ce6e84844b55', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'f5854cfd-f990-46fd-a6cb-77b48a80b24b'),
('15796145-1534-4362-977e-69699646cda0', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '6f4fe3df-9ab8-46c1-a8b7-9458413980f3'),
('6cd703ce-72b6-466b-8262-ba3fdf1d1d57', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '9c644140-c71d-4435-bd6b-a81171913046'),
('39a58779-f4c3-4e46-b9af-3ae867554d2e', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '259fb434-5321-43be-b007-76f3c2dfbfcc'),
('cad5bfb9-2421-487b-a217-e8e2108271c7', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'fa6961f1-1cee-494f-8f1f-f4552b49b6fa'),
('a5c9bbc2-5ab6-4219-add8-532763ff9490', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '60768a38-8ff1-493a-85ba-b62df98a60d4'),
('e3db8e5a-0f9d-4ab9-8d6f-f16e3e3f8bbe', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'a66215da-48dc-419d-8e2d-3f409aa038e3'),
('9b175068-7a0f-415c-8e98-5fa92895984f', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '36e9c7bd-c98f-4414-a8f4-164da5299874'),
('e84db5f8-66bc-408c-a38b-659cdf254b0b', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'ca0af093-65ed-48d4-a6d1-62985a35fd8f'),
('4dcc0416-d5e8-43a7-8694-2cd55c2f160e', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'd93ede75-8fd5-4f88-b119-07493b757f4b'),
('10649145-3440-49f4-bb42-b6670eabd579', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '6118a1da-aa9a-40eb-b664-87a6969c5759'),
('4177d9fb-9968-4eae-b3e3-66541295305a', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '79b1a377-7b73-4bea-90c2-330bda11b635'),
('0f987cdd-1af4-4d54-bd5f-017ce5a6f64d', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', 'e33cc269-3a12-436e-aab7-75f14953388f'),
('09c991c4-bb65-4088-badb-80f916a51600', NULL, '2022-11-11 18:35:27.069944', '2022-11-11 18:35:27.069944', '11ccffea-2cc9-4e98-9bef-3464dfe4dec8', '2f6c957c-725b-4da4-ad6d-9a23a1e18a98');

INSERT INTO public.positions (id, deleted_at, created_at, updated_at, name, code) VALUES
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Frontend', 'frontend'),
('d796884d-a8c4-4525-81e7-54a3b6099eac', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Backend', 'backend'),
('dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Devops', 'devops'),
('01fb6322-d727-47e3-a242-5039ea4732fc', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Blockchain', 'blockchain'),
('39735742-829b-47f3-8f9d-daf0983914e5', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Project-Management', 'project-management');

INSERT INTO public.countries (id, deleted_at, created_at, updated_at, name, code, cities) VALUES
('da9031ce-0d6e-4344-b97a-a2c44c66153e', null, '2022-11-08 08:08:09.881727', '2022-11-08 08:08:09.881727', 'Singapore', '+65', '[{"lat": "", "long": "", "name": "Singapore"}]'),
('4ef64490-c906-4192-a7f9-d2221dadfe4c', null, '2022-11-08 08:06:56.068148', '2022-11-08 08:06:56.068148', 'Vietnam', '+84', '[{"lat": "", "long": "", "name": "Hồ Chí Minh"}, {"lat": "", "long": "", "name": "An Giang"}, {"lat": "", "long": "", "name": "Bà Rịa-Vũng Tàu"}, {"lat": "", "long": "", "name": "Bình Dương"}, {"lat": "", "long": "", "name": "Bình Định"}, {"lat": "", "long": "", "name": "Bình Phước"}, {"lat": "", "long": "", "name": "Bình Thuận"}, {"lat": "", "long": "", "name": "Bạc Liêu"}, {"lat": "", "long": "", "name": "Bắc Giang"}, {"lat": "", "long": "", "name": "Bắc Kạn"}, {"lat": "", "long": "", "name": "Bắc Ninh"}, {"lat": "", "long": "", "name": "Bến Tre"}, {"lat": "", "long": "", "name": "Cao Bằng"}, {"lat": "", "long": "", "name": "Cà Mau"}, {"lat": "", "long": "", "name": "Cần Thơ"}, {"lat": "", "long": "", "name": "Điện Biên"}, {"lat": "", "long": "", "name": "Đà Nẵng"}, {"lat": "", "long": "", "name": "Đắk Lắk"}, {"lat": "", "long": "", "name": "Đồng Nai"}, {"lat": "", "long": "", "name": "Đắk Nông"}, {"lat": "", "long": "", "name": "Đồng Tháp"}, {"lat": "", "long": "", "name": "Gia Lai"}, {"lat": "", "long": "", "name": "Hoà Bình"}, {"lat": "", "long": "", "name": "Hà Giang"}, {"lat": "", "long": "", "name": "Hà Nam"}, {"lat": "", "long": "", "name": "Hà Nội"}, {"lat": "", "long": "", "name": "Hà Tĩnh"}, {"lat": "", "long": "", "name": "Hải Dương"}, {"lat": "", "long": "", "name": "Hải Phòng"}, {"lat": "", "long": "", "name": "Hậu Giang"}, {"lat": "", "long": "", "name": "Hưng Yên"}, {"lat": "", "long": "", "name": "Khánh Hòa"}, {"lat": "", "long": "", "name": "Kiên Giang"}, {"lat": "", "long": "", "name": "Kon Tum"}, {"lat": "", "long": "", "name": "Lai Châu"}, {"lat": "", "long": "", "name": "Lâm Đồng"}, {"lat": "", "long": "", "name": "Lạng Sơn"}, {"lat": "", "long": "", "name": "Lào Cai"}, {"lat": "", "long": "", "name": "Long An"}, {"lat": "", "long": "", "name": "Nam Định"}, {"lat": "", "long": "", "name": "Nghệ An"}, {"lat": "", "long": "", "name": "Ninh Bình"}, {"lat": "", "long": "", "name": "Ninh Thuận"}, {"lat": "", "long": "", "name": "Phú Thọ"}, {"lat": "", "long": "", "name": "Phú Yên"}, {"lat": "", "long": "", "name": "Quảng Bình"}, {"lat": "", "long": "", "name": "Quảng Nam"}, {"lat": "", "long": "", "name": "Quảng Ngãi"}, {"lat": "", "long": "", "name": "Quảng Ninh"}, {"lat": "", "long": "", "name": "Quảng Trị"}, {"lat": "", "long": "", "name": "Sóc Trăng"}, {"lat": "", "long": "", "name": "Sơn La"}, {"lat": "", "long": "", "name": "Thanh Hóa"}, {"lat": "", "long": "", "name": "Thái Bình"}, {"lat": "", "long": "", "name": "Thái Nguyên"}, {"lat": "", "long": "", "name": "Thừa Thiên Huế"}, {"lat": "", "long": "", "name": "Tiền Giang"}, {"lat": "", "long": "", "name": "Trà Vinh"}, {"lat": "", "long": "", "name": "Tuyên Quang"}, {"lat": "", "long": "", "name": "Tây Ninh"}, {"lat": "", "long": "", "name": "Vĩnh Long"}, {"lat": "", "long": "", "name": "Vĩnh Phúc"}, {"lat": "", "long": "", "name": "Yên Bái"}]');

INSERT INTO public.seniorities (id, deleted_at, created_at, updated_at, name, code, level) VALUES
('11ccffea-2cc9-4e98-9bef-3464dfe4dec8', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Fresher', 'fresher', 1),
('d796884d-a8c4-4525-81e7-54a3b6099eac', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Junior', 'junior', 2),
('dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Mid', 'mid', 3),
('01fb6322-d727-47e3-a242-5039ea4732fc', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Senior', 'senior', 4),
('01fb6322-d727-47e3-a242-5039ea4732fd', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Staff', 'staff', 5),
('39735742-829b-47f3-8f9d-daf0983914e5', NULL, '2022-11-07 09:50:25.714604', '2022-11-07 09:50:25.714604', 'Principal', 'principal', 6);

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
('7d95e035-81d6-49d7-bed4-3a83bf2e34d6', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'How effective of a leader is this person, either through direct management or influence?', 'general', 2, NULL),
('d36e84c5-d7a4-4d5f-ada1-f6b9ddb58f51', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'Does this person find creative solutions, and own the solution to problems? Are they proactive or reactive?', 'general', 3, NULL),
('f03432ba-c024-467e-8059-a5bb2b7f783d', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'How would you rate the quality of the employee''s work?', 'general', 4, NULL),
('d2bb48c1-e8d6-4946-a372-8499907b7328', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'How well does this person set and meet deadlines?', 'general', 5, NULL),
('be86ce52-803b-403f-b059-1a69492fe3d4', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'How well does this person embody our culture?', 'general', 6, NULL),
('51eab8c7-61ba-4c56-be39-b72eb6b89a52', NULL, '2022-12-06 03:02:39.049420', '2022-12-06 03:02:39.049420', 'survey', 'peer-review', 'If you could give this person one piece of constructive advice to make them more effective in their role, what would you say?', 'general', 7, NULL),
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

INSERT INTO public.clients (id, deleted_at, created_at, updated_at, name, description, registration_number, address, country, industry, website, emails) VALUES
('afb9cf05-9517-4fb9-a4f2-66e6d90ad215', null, '2023-02-07 18:41:35.740901', '2023-02-07 18:41:35.740901', 'Lorem Ipsum Inc.', 'Lorem ipsum dolor sit amet, consectetur adipiscing elit.', '32320398488', 'Hado Centrosa', 'Vietnam', 'Technology', 'https://d.foundations', '{benjamin@d.foundation,namnh@d.foundation}');

INSERT INTO public.client_contacts (id, deleted_at, created_at, updated_at, name, client_id, role, metadata, emails, is_main_contact) VALUES
('0569e64d-3b57-454a-ab88-0482e087eb5f', null, '2023-02-07 19:05:04.354072', '2023-02-07 19:05:04.354072', 'Thanh Pham ', 'afb9cf05-9517-4fb9-a4f2-66e6d90ad215', 'PM', null, '["thanh@d.foundation", "huytq@d.foundation"]', true);

INSERT INTO public.company_infos (id, deleted_at, created_at, updated_at, name, description, registration_number, info) VALUES
('2b57ec32-19c2-46f0-8cf5-04623241a464', null, '2023-02-07 18:42:38.328707', '2023-02-07 18:42:38.328707', 'Dwarves Foundation', null, '1245888282', '{"vn": {"phone": "0988999999", "address": "Hado Centrosa"}}');

INSERT INTO public.invoice_number_caching (id, deleted_at, created_at, updated_at, key, number) VALUES
('25938962-611b-45ad-b6be-c0a1365ea1de', null, '2023-02-07 18:57:47.312554', '2023-02-07 18:57:47.312554', 'year_invoice_2023', 1),
('10fa5046-199d-4d9a-bbf9-44345127be79', null, '2023-02-07 18:59:45.024935', '2023-02-07 18:59:45.024935', 'project_invoice_fortress_2023', 1);

INSERT INTO public.bank_accounts (id, deleted_at, created_at, updated_at, account_number, bank_name, currency_id, owner_name, address, swift_code, routing_number, name, uk_sort_code) VALUES
('fc6b1743-05c5-4152-9340-1d20d96d8fc0', null, '2023-02-07 18:39:35.547782', '2023-02-07 18:39:35.547782', '0999999888', 'ACB', '7037bdb6-584e-4e35-996d-ef28a243f48a', 'Dwarves Foundation', 'Hado Centrosa', 'AVBWFPW', null, 'DF Bank Account', null);

INSERT INTO public.organizations (id, deleted_at, created_at, updated_at, name, code, avatar) VALUES
('31fdf38f-77c0-4c06-b530-e2be8bc297e0', NULL, '2023-01-19 11:13:13.487168', '2023-01-19 11:13:13.487168', 'Dwarves Foundation', 'dwarves-foundation', NULL),
('e4725383-943a-468a-b0cd-ce249c573cf7', NULL, '2023-01-19 11:13:13.487168', '2023-01-19 11:13:13.487168', 'Console Labs', 'console-labs', NULL);

