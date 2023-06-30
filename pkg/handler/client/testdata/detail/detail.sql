INSERT INTO public.projects (id, deleted_at, created_at, updated_at, name, type, start_date, end_date, status, country_id, client_email, project_email) VALUES
('dfa182fc-1d2d-49f6-a877-c01da9ce4207', NULL, '2022-11-11 18:06:56.362902', '2022-11-11 18:06:56.362902', 'Lorem ipsum', 'time-material', '2022-07-06', NULL, 'active', NULL, NULL, NULL),
('8dc3be2e-19a4-4942-8a79-56db391a0b15', NULL, '2022-11-11 18:06:56.362902', '2022-11-11 18:06:56.362902', 'Fortress', 'dwarves', '2022-11-01', NULL, 'active', '4ef64490-c906-4192-a7f9-d2221dadfe4c', 'team@d.foundation', 'fortress@d.foundation');

INSERT INTO public.project_slots (id, deleted_at, created_at, updated_at, project_id, seniority_id, upsell_person_id, deployment_type, rate, discount, status) VALUES
('f32d08ca-8863-4ab3-8c84-a11849451eb7', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL, 'official', 5000, 0, 'active'),
('bdc64b18-4c5f-4025-827a-f5b91d599dc7', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL, 'shadow', 4000, 0, 'active'),
('1406fcce-6f90-4e0f-bea1-c373e2b2b5b1', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL, 'official', 3000, 3, 'active'),
('b25bd3fa-eb6d-49d5-b278-7aacf4594f79', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL, 'official', 3000, 3, 'active'),
('ce379dc0-95be-471a-9227-8e045a5630af', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e', NULL, 'shadow', 4000, 0, 'active');

INSERT INTO public.project_members (id, deleted_at, created_at, updated_at, project_id, project_slot_id ,employee_id, seniority_id, start_date, end_date, rate, discount, status,deployment_type, upsell_person_id) VALUES
('cb889a9c-b20c-47ee-83b8-44b6d1721aca', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'f32d08ca-8863-4ab3-8c84-a11849451eb7', '2655832e-f009-4b73-a535-64c3a22e558f', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', NULL, 5000, 0, 'active', 'official', '608ea227-45a5-4c8a-af43-6c7280d96340'),
('7310b51a-3855-498b-99ab-41aa82934269', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'bdc64b18-4c5f-4025-827a-f5b91d599dc7', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', NULL, 4000, 0, 'active', 'shadow', '608ea227-45a5-4c8a-af43-6c7280d96340'),
('35149aab-0506-4eb3-9300-c706ccbf2bde', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '1406fcce-6f90-4e0f-bea1-c373e2b2b5b1', '8d7c99c0-3253-4286-93a9-e7554cb327ef', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', NULL, 3000, 3, 'active', 'official', '608ea227-45a5-4c8a-af43-6c7280d96340'),
('5a9a07aa-e8f3-4b62-b9ad-0f057866dc6c', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'b25bd3fa-eb6d-49d5-b278-7aacf4594f79', 'eeae589a-94e3-49ac-a94c-fcfb084152b2', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', NULL, 3000, 3, 'active', 'official', '608ea227-45a5-4c8a-af43-6c7280d96340'),
('fcd5c16f-40fd-48b6-9649-410691373eea', NULL, '2022-11-11 18:19:56.156172', '2022-11-11 18:19:56.156172', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'ce379dc0-95be-471a-9227-8e045a5630af', '608ea227-45a5-4c8a-af43-6c7280d96340', 'dac16ce6-9e5a-4ff3-9ea2-fdea4853925e' ,'2022-11-01', NULL, 4000, 0, 'active', 'shadow', '608ea227-45a5-4c8a-af43-6c7280d96340');

INSERT INTO public.project_heads (id, deleted_at, created_at, updated_at, project_id, employee_id, start_date, end_date, commission_rate, position) VALUES
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

INSERT INTO public.work_units (id, deleted_at, created_at, updated_at, name, status, type, source_url, project_id, source_metadata) VALUES 
('4797347d-21e0-4dac-a6c7-c98bf2d6b27c', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'Fortress API', 'active', 'development', 'https://github.com/dwarvesf/fortress-api', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '[]'),
('69b32f7e-0433-4566-a801-72909172940e', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'Fortress Web', 'archived', 'management', 'https://github.com/dwarvesf/fortress-web', 'dfa182fc-1d2d-49f6-a877-c01da9ce4207', '[]');

INSERT INTO public.work_unit_stacks (id, deleted_at, created_at, updated_at, stack_id, work_unit_id) VALUES 
('95851b98-c8d0-46f6-b6b0-8dd2037f44d6', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', '0ecf47c8-cca4-4c30-94bb-054b1124c44f', '4797347d-21e0-4dac-a6c7-c98bf2d6b27c'),
('f1ddeeb2-ad44-4c97-a934-86ad8f24ca57', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'fa0f4e46-7eab-4e5c-9d31-30489e69fe2e', '4797347d-21e0-4dac-a6c7-c98bf2d6b27c'),
('9b7c9d01-75a3-4386-93a7-4ff099887847', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'b403ef95-4269-4830-bbb6-8e56e5ec0af4', '69b32f7e-0433-4566-a801-72909172940e'),
('b851f3bc-a758-4e28-834b-0d2c0a04bf71', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', 'fa0f4e46-7eab-4e5c-9d31-30489e69fe2e', '69b32f7e-0433-4566-a801-72909172940e');

INSERT INTO public.work_unit_members (id, deleted_at, created_at, updated_at, start_date, end_date, status, project_id, employee_id, work_unit_id) VALUES 
('303fd2e5-0b4d-401c-b5fa-74820991e6c0', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', '2022-11-29', NULL, 'active', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '2655832e-f009-4b73-a535-64c3a22e558f', '4797347d-21e0-4dac-a6c7-c98bf2d6b27c'),
('f79ae054-4ab4-41cd-aa5f-c871887cc35c', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', '2022-11-29', NULL, 'active', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', '69b32f7e-0433-4566-a801-72909172940e'),
('7e4da4ac-241f-4af8-b0a0-f59e5a64065b', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', '2022-11-29', NULL, 'active', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '8d7c99c0-3253-4286-93a9-e7554cb327ef', '69b32f7e-0433-4566-a801-72909172940e'),
('93954bcd-d5e9-4c4c-ad30-7de5fd332a80', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', '2022-11-29', NULL, 'inactive', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'eeae589a-94e3-49ac-a94c-fcfb084152b2', '69b32f7e-0433-4566-a801-72909172940e'),
('e14f68f8-7ed5-4a59-9df8-275573537861', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', '2022-11-29', NULL, 'inactive', '8dc3be2e-19a4-4942-8a79-56db391a0b15', '608ea227-45a5-4c8a-af43-6c7280d96340', '4797347d-21e0-4dac-a6c7-c98bf2d6b27c'),
('799708d7-855b-4a42-b169-7c9891f0b218', NULL, '2022-11-29 08:03:33.233262', '2022-11-29 08:03:33.233262', '2022-11-29', NULL, 'inactive', '8dc3be2e-19a4-4942-8a79-56db391a0b15', 'ecea9d15-05ba-4a4e-9787-54210e3b98ce', '4797347d-21e0-4dac-a6c7-c98bf2d6b27c');

INSERT INTO public.clients (id, deleted_at, created_at, updated_at, name, description, registration_number, address, country, industry, website, emails) VALUES 
('377241ad-1216-46be-9733-fa624da60555', null, '2023-02-10 16:21:03.774651', '2023-02-10 16:21:03.774651', 'name', 'description', '123', 'LA', 'USA', 'technical', 'a.com', null),
('67f9f420-cdd5-4793-88c7-d2068bd17f61', null, '2023-02-10 16:40:03.932475', '2023-02-10 16:40:03.932475', 'John', 'description', '456', 'CA', 'USA', 'education', 'b.com', null);

INSERT INTO public.client_contacts (id, deleted_at, created_at, updated_at, name, client_id, role, metadata, emails, is_main_contact) VALUES 
('c191a78c-0d34-4bc8-bbb6-18ffa63750d9', null, '2023-02-10 16:21:03.777661', '2023-02-10 16:21:03.777661', 'contact name', '377241ad-1216-46be-9733-fa624da60555', 'manager', null, '{"emails": ["a@gmail.com", "b@gmail.com"]}', true),
('596ae21d-570b-4852-92a6-117633629046', null, '2023-02-10 16:40:03.934235', '2023-02-10 16:40:03.934235', 'contact name 1', '67f9f420-cdd5-4793-88c7-d2068bd17f61', 'manager', null, '{"emails": ["john1@gmail.com", "john2@gmail.com"]}', true),
('bebfe6b3-a09b-4a90-bba5-fe8a97d00047', null, '2023-02-10 16:40:03.935914', '2023-02-10 16:40:03.935914', 'contact name 2', '67f9f420-cdd5-4793-88c7-d2068bd17f61', 'manager', null, '{"emails": ["john3@gmail.com", "john4@gmail.com"]}', false);
