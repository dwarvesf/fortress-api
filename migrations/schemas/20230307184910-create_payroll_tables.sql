
-- +migrate Up
UPDATE "employees" SET "basecamp_id" = NULL WHERE "basecamp_id" = '';
ALTER TABLE "employees" ALTER COLUMN "basecamp_id" TYPE integer USING ("basecamp_id"::integer);

CREATE TABLE IF NOT EXISTS "base_salaries" (
  "id" uuid PRIMARY KEY DEFAULT (uuid()),
  "employee_id" uuid,
  "contract_amount" int8 NOT NULL DEFAULT 0,
  "company_account_amount" int8 NOT NULL DEFAULT 0,
  "personal_account_amount" int8 NOT NULL DEFAULT 0,
  "insurance_amount" int8 NOT NULL DEFAULT 0,
  "currency_id" uuid,
  "effective_date" date,
  "created_at" timestamptz(6) DEFAULT now(),
  "deleted_at" timestamptz(6),
  "is_active" bool DEFAULT true,
  "batch" int4,
  "type" text COLLATE "pg_catalog"."default",
  "category" text COLLATE "pg_catalog"."default"
)
;
ALTER TABLE "base_salaries" ADD CONSTRAINT "base_salaries_currency_id_fkey" FOREIGN KEY ("currency_id") REFERENCES "currencies" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "base_salaries" ADD CONSTRAINT "base_salaries_employee_id_fkey" FOREIGN KEY ("employee_id") REFERENCES "employees" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

CREATE TABLE IF NOT EXISTS  "accounting_transactions" (
  "id" uuid PRIMARY KEY DEFAULT (uuid()),
  "created_at" timestamptz(6) DEFAULT now(),
  "deleted_at" timestamptz(6),
  "date" timestamptz(6) DEFAULT now(),
  "name" text COLLATE "pg_catalog"."default",
  "amount" float8,
  "currency_id" uuid NOT NULL,
  "conversion_amount" int8,
  "organization" text COLLATE "pg_catalog"."default",
  "metadata" json,
  "category" text COLLATE "pg_catalog"."default",
  "currency" text COLLATE "pg_catalog"."default",
  "conversion_rate" float4,
  "type" text COLLATE "pg_catalog"."default"
)
;
ALTER TABLE "accounting_transactions" ADD CONSTRAINT "transaction_info_unique" UNIQUE ("name", "date");
ALTER TABLE "accounting_transactions" DROP CONSTRAINT IF EXISTS "accounting_transactions_currency_id_fkey";
ALTER TABLE "accounting_transactions" ADD CONSTRAINT "accounting_transactions_currency_id_fkey" FOREIGN KEY ("currency_id") REFERENCES "currencies" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

CREATE TABLE IF NOT EXISTS "accounting_categories" (
    "id" uuid PRIMARY KEY DEFAULT (uuid()),
    "created_at" timestamptz(6) DEFAULT now(),
    "updated_at" TIMESTAMPTZ(6) DEFAULT NOW(),
    "deleted_at" timestamptz(6),
    "name" text COLLATE "pg_catalog"."default",
    "type" text COLLATE "pg_catalog"."default"
);

CREATE TABLE IF NOT EXISTS "employee_commissions" (
    "id" UUID PRIMARY KEY DEFAULT (uuid()),
    "created_at" TIMESTAMPTZ(6) DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ(6) DEFAULT NOW(),
    "deleted_at" TIMESTAMPTZ(6),
    "invoice_id" UUID NOT NULL,
    "employee_id" UUID,
    "amount" INT4,
    "project" TEXT COLLATE "pg_catalog"."default",
    "conversion_rate" DECIMAL DEFAULT 0,
    "is_paid" BOOL DEFAULT FALSE,
    "formula" TEXT COLLATE "pg_catalog"."default",
    "note" TEXT COLLATE "pg_catalog"."default",
    "paid_at" TIMESTAMPTZ(6)
);
ALTER TABLE "employee_commissions" ADD CONSTRAINT "employee_commissions_employee_id_fkey" FOREIGN KEY ("employee_id") REFERENCES "employees" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

CREATE TABLE IF NOT EXISTS "employee_bonuses" (
  "id" uuid PRIMARY KEY DEFAULT (uuid()),
  "employee_id" uuid NOT NULL,
  "amount" int8,
  "is_active" bool DEFAULT true,
  "name" text COLLATE "pg_catalog"."default",
  "created_at" timestamp(6) DEFAULT now()
)
;
ALTER TABLE "employee_bonuses" ADD CONSTRAINT "employee_bonuses_employee_id_fkey" FOREIGN KEY ("employee_id") REFERENCES "employees" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

CREATE TABLE IF NOT EXISTS "payrolls" (
    "id" uuid PRIMARY KEY DEFAULT (uuid()),
    "employee_id" uuid NOT NULL,
    "total" int8 NOT NULL DEFAULT 0,
    "month" int4,
    "year" int4,
    "commission_amount" int8 NOT NULL DEFAULT 0,
    "commission_explain" json,
    "employee_rank_snapshot" json,
    "total_explain" json,
    "project_bonus_amount" int8 NOT NULL DEFAULT 0,
    "due_date" date,
    "project_bonus_explain" json,
    "is_paid" bool DEFAULT false,
    "conversion_amount" int8 NOT NULL DEFAULT 0,
    "base_salary_amount" int8 NOT NULL DEFAULT 0,
    "contract_amount" int8
);
ALTER TABLE
    "payrolls"
ADD
    CONSTRAINT "payrolls_employee_id_fkey" FOREIGN KEY ("employee_id") REFERENCES "employees" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

CREATE TABLE IF NOT EXISTS "project_commission_configs" (
    "id" UUID PRIMARY KEY DEFAULT (uuid()),
    "deleted_at" TIMESTAMP(6),
    "created_at" TIMESTAMP(6) DEFAULT (NOW()),
    "updated_at" TIMESTAMP(6) DEFAULT (NOW()),
    "project_id" UUID,
    "position" project_head_positions,
    "commission_rate" DECIMAL
);

ALTER TABLE
    "project_commission_configs"
ADD
    CONSTRAINT "project_commission_configs_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "projects" ("id");

CREATE TABLE IF NOT EXISTS "cached_payrolls" (
  "id" UUID PRIMARY KEY DEFAULT (uuid()),
  "month" int4,
  "year" int4,
  "batch" int4,
  "payrolls" json
)
;
ALTER TABLE "cached_payrolls" ADD CONSTRAINT "cached_payrolls_month_year_batch_key" UNIQUE ("month", "year", "batch");

-- +migrate Down
