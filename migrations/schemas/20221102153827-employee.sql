
-- +migrate Up
CREATE or replace FUNCTION  uuid() RETURNS uuid
    LANGUAGE c
    AS '$libdir/pgcrypto', 'pg_random_uuid';

CREATE TYPE employment_status as ENUM('left', 'probation', 'full-time', 'contractor');

create table if not exists employees (
  id uuid primary key DEFAULT uuid(),
  deleted_at timestamp(6),
  created_at timestamp(6) default now(),
  updated_at timestamp(6) default now(),
  full_name text not null,
  display_name text not null,
  gender text not null,
  team_email text not null,
  personal_email text not null,
  avatar text not null,
  phone_number text,
  address text,
  mbti text,
  horoscope text,
  passport_photo_front text,
  passport_photo_back text,
  identity_card_photo_front text,
  identity_card_photo_back text,
  date_of_birth date,
  employment_status employment_status,
  joined_date date,
  left_date date,
  basecamp_id text,
  basecamp_attachable_sgid text,
  gitlab_id text,
  github_id text,
  discord_id text,
  wise_recipient_email text,
  wise_recipient_name text,
  wise_recipient_id text,
  wise_account_number text,
  wise_currency text,
  local_bank_branch text,
  local_bank_number text,
  local_bank_currency text,
  local_branch_name text,
  local_bank_recipient_name text
);

-- +migrate Down
drop table employees;
drop type employment_status;
