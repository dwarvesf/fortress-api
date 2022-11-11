-- +migrate Up
CREATE OR REPLACE FUNCTION  uuid() RETURNS uuid
    LANGUAGE c
    AS '$libdir/pgcrypto', 'pg_random_uuid';

CREATE TYPE employment_status as ENUM('left', 'probation', 'full-time', 'contractor');

CREATE TABLE IF NOT EXISTS employees (
  id uuid PRIMARY KEY DEFAULT uuid(),
  deleted_at TIMESTAMP(6),
  created_at TIMESTAMP(6) DEFAULT NOW(),
  updated_at TIMESTAMP(6) DEFAULT NOW(),
  full_name TEXT NOT NULL,
  display_name TEXT NOT NULL,
  gender TEXT NOT NULL,
  team_email TEXT NOT NULL,
  personal_email TEXT NOT NULL,
  avatar TEXT NOT NULL,
  phone_number TEXT,
  address TEXT,
  mbti TEXT,
  horoscope TEXT,
  passport_photo_front TEXT,
  passport_photo_back TEXT,
  identity_card_photo_front TEXT,
  identity_card_photo_back TEXT,
  date_of_birth date,
  employment_status employment_status,
  joined_date date,
  left_date date,
  basecamp_id TEXT,
  basecamp_attachable_sgid TEXT,
  gitlab_id TEXT,
  github_id TEXT,
  discord_id TEXT,
  wise_recipient_email TEXT,
  wise_recipient_name TEXT,
  wise_recipient_id TEXT,
  wise_account_number TEXT,
  wise_currency TEXT,
  local_bank_branch TEXT,
  local_bank_number TEXT,
  local_bank_currency TEXT,
  local_branch_name TEXT,
  local_bank_recipient_name TEXT
);

-- +migrate Down
DROP TABLE IF EXISTS employees;
DROP TYPE IF EXISTS employment_status;
