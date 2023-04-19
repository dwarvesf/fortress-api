
-- +migrate Up
alter table public.bank_accounts add column intermediary_bank_address text;
alter table public.bank_accounts add column intermediary_bank_name text;

-- +migrate Down
alter table public.bank_accounts drop column intermediary_bank_address;
alter table public.bank_accounts drop column intermediary_bank_name;
