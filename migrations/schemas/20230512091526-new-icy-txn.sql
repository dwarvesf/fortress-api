-- +migrate Up

-- +migrate StatementBegin
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'enum_icy_txn_category') THEN
    CREATE TYPE "enum_icy_txn_category" AS ENUM ('learning', 'community', 'delivery', 'tooling');
  END IF;
END $$;
-- +migrate StatementEnd

CREATE TABLE IF NOT EXISTS "icy_transactions" (
  "id" uuid PRIMARY KEY DEFAULT uuid(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz,
  "txn_time" timestamptz NOT NULL DEFAULT now(),
  "src_employee_id" uuid REFERENCES employees ("id"),
  "dest_employee_id" uuid REFERENCES employees ("id"),
  "category" enum_icy_txn_category NOT NULL,
  "amount" numeric NOT NULL DEFAULT 0,
  "note" text
);

ALTER TABLE
  audiences
ADD
  column unsub_at TIMESTAMP default null;

-- +migrate Down

DROP TABLE icy_transactions;

DROP TYPE IF EXISTS "enum_icy_txn_category";
