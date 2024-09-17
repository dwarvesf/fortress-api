-- +migrate Up
ALTER TABLE employees ADD COLUMN beneficiary_bank_name TEXT DEFAULT NULL;
ALTER TABLE employees ADD COLUMN beneficiary_bank_postcode TEXT DEFAULT NULL;
ALTER TABLE employees ADD COLUMN beneficiary_bank_address TEXT DEFAULT NULL;
ALTER TABLE employees ADD COLUMN beneficiary_bank_city TEXT DEFAULT NULL;
ALTER TABLE employees ADD COLUMN beneficiary_routing_number TEXT DEFAULT NULL;

-- +migrate Down
ALTER TABLE employees DROP COLUMN beneficiary_bank_name;
ALTER TABLE employees DROP COLUMN beneficiary_bank_postcode;
ALTER TABLE employees DROP COLUMN beneficiary_bank_address;
ALTER TABLE employees DROP COLUMN beneficiary_bank_city;
ALTER TABLE employees DROP COLUMN beneficiary_routing_number;
