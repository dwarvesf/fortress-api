-- +migrate Up
-- Add JSONB column to store discord account IDs
ALTER TABLE memo_logs 
ADD COLUMN discord_account_ids JSONB DEFAULT '[]'::JSONB;

-- Migrate data from memo_authors to memo_logs
WITH author_data AS (
  SELECT 
    memo_log_id, 
    json_agg(discord_account_id) AS discord_account_ids
  FROM memo_authors
  GROUP BY memo_log_id
)
UPDATE memo_logs ml
SET discord_account_ids = ad.discord_account_ids
FROM author_data ad
WHERE ml.id = ad.memo_log_id;

-- Drop the memo_authors table
DROP TABLE memo_authors;
DROP TABLE brainery_logs;

-- +migrate Down
CREATE TABLE "brainery_logs" (
    "id" uuid NOT NULL DEFAULT uuid(),
    "deleted_at" timestamp,
    "created_at" timestamp DEFAULT now(),
    "updated_at" timestamp DEFAULT now(),
    "title" text NOT NULL,
    "url" text NOT NULL,
    "github_id" text,
    "discord_id" text NOT NULL,
    "employee_id" uuid,
    "tags" jsonb,
    "published_at" timestamp NOT NULL,
    "reward" numeric,
    CONSTRAINT "brainery_logs_employee_id_fkey" FOREIGN KEY ("employee_id") REFERENCES "employees"("id"),
    PRIMARY KEY ("id")
);
CREATE TABLE IF NOT EXISTS memo_authors (
  memo_log_id UUID NOT NULL REFERENCES memo_logs(id),
  discord_account_id UUID NOT NULL REFERENCES discord_accounts(id),
  created_at TIMESTAMP(6) DEFAULT (now()),
  PRIMARY KEY (memo_log_id, discord_account_id)
);

INSERT INTO memo_authors (memo_log_id, discord_account_id)
SELECT 
  id AS memo_log_id, 
  jsonb_array_elements(discord_account_ids)::UUID AS discord_account_id
FROM memo_logs
WHERE jsonb_array_length(discord_account_ids) > 0;

-- Remove the JSONB column
ALTER TABLE memo_logs 
DROP COLUMN discord_account_ids;
