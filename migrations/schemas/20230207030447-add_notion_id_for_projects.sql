
-- +migrate Up
-- +migrate StatementBegin
UPDATE projects SET notion_id = '62dc0545-9548-4761-9a62-625cad4f4246' WHERE id = 'dc8612ca-e46d-4a69-a58f-34272351ff06';
UPDATE projects SET notion_id = '72f64a6c-1be0-474c-a40c-0c3091e5be1e' WHERE id = '4f7bd4e6-f8d1-45f3-886c-82518079d47b';
UPDATE projects SET notion_id = '2df3256d-ea4a-45ed-a91b-4b3d2a7250e2' WHERE id = '8afe840a-80fa-48fb-aa26-c6c9a55ba983';
UPDATE projects SET notion_id = '672a9da4-d037-4fde-821e-7d7a4bc0770a' WHERE id = '1ed315b4-12b2-49f8-a6fe-b1ddc4659eb8';
UPDATE projects SET notion_id = '22f925ad-9788-4deb-a7fa-420b92651a02' WHERE id = 'ebb60490-a766-4d6c-9871-46857872799e';
UPDATE projects SET notion_id = '81ffc0a7-52ee-43fd-92ae-8603b02008e8' WHERE id = 'bf9181f4-86cb-4773-aede-9f7be28f4bd1';
UPDATE projects SET notion_id = '2907613b-a86e-4b20-b8c6-4405a590fff7' WHERE id = 'fa1127b4-c802-4b6d-8cce-5e978efdab29';
UPDATE projects SET notion_id = 'c59cd991-809e-4bcf-a9b5-ee46d61217d7' WHERE id = '516d8f9d-5e60-497c-a244-6ab06170e5c3';
UPDATE projects SET notion_id = 'bfabfc8f-5ba7-4f9e-bade-772a7ad05109' WHERE id = '1ed315b5-12b2-49f8-a6fe-b1ddc4659eb8';
UPDATE projects SET notion_id = '87d91bf6-1301-4bbd-aca4-2a9234da9510' WHERE id = 'adcccfdb-0dff-462f-9916-44ca94d4cf73';
UPDATE projects SET notion_id = '3938e825-4e49-4bb5-be64-4033bf21990b' WHERE id = 'fd619ddf-79df-4ee1-b96e-a6fdb927450e';
UPDATE projects SET notion_id = '1ed54c0e-39b4-4c2f-bbb2-f466b64f77c1' WHERE id = 'cab8e2ec-f78b-4290-8083-1e1d64c49165';
UPDATE projects SET notion_id = '17765181-8089-400f-8c3a-5aa61184337c' WHERE id = 'f5aaa264-fbcd-4087-8897-7b2f5048991c';
-- +migrate StatementEnd
-- +migrate Down
SELECT TRUE;
