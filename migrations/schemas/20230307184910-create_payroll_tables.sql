-- +migrate Up
UPDATE "employees"
SET "basecamp_id" = NULL
WHERE "basecamp_id" = '';
ALTER TABLE "employees"
ALTER COLUMN "basecamp_id" TYPE integer USING ("basecamp_id"::integer);

CREATE TABLE IF NOT EXISTS "base_salaries" (
    "id"                      UUID PRIMARY KEY DEFAULT (uuid()),
    "created_at"              TIMESTAMPTZ(6)   DEFAULT NOW(),
    "deleted_at"              TIMESTAMPTZ(6),

    "employee_id"             UUID,
    "contract_amount"         int8 NOT NULL    DEFAULT 0,
    "company_account_amount"  int8 NOT NULL    DEFAULT 0,
    "personal_account_amount" int8 NOT NULL    DEFAULT 0,
    "insurance_amount"        int8 NOT NULL    DEFAULT 0,
    "currency_id"             UUID,
    "effective_date"          DATE,
    "is_active"               BOOL             DEFAULT TRUE,
    "batch"                   INT4,
    "type"                    TEXT COLLATE "pg_catalog"."default",
    "category"                TEXT COLLATE "pg_catalog"."default"
);
ALTER TABLE "base_salaries"
    ADD CONSTRAINT "base_salaries_currency_id_fkey" FOREIGN KEY ("currency_id") REFERENCES "currencies" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "base_salaries"
    ADD CONSTRAINT "base_salaries_employee_id_fkey" FOREIGN KEY ("employee_id") REFERENCES "employees" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
CREATE TABLE IF NOT EXISTS "accounting_categories" (
    "id"         UUID PRIMARY KEY DEFAULT (uuid()),
    "created_at" TIMESTAMPTZ(6)   DEFAULT now(),
    "updated_at" TIMESTAMPTZ(6)   DEFAULT NOW(),
    "deleted_at" TIMESTAMPTZ(6),

    "name"       TEXT COLLATE "pg_catalog"."default",
    "type"       TEXT COLLATE "pg_catalog"."default"
);

CREATE TABLE IF NOT EXISTS "employee_commissions" (
    "id"              UUID PRIMARY KEY DEFAULT (uuid()),
    "created_at"      TIMESTAMPTZ(6)   DEFAULT NOW(),
    "updated_at"      TIMESTAMPTZ(6)   DEFAULT NOW(),
    "deleted_at"      TIMESTAMPTZ(6),

    "invoice_id"      UUID NOT NULL,
    "employee_id"     UUID,
    "amount"          INT4,
    "project"         TEXT COLLATE "pg_catalog"."default",
    "conversion_rate" DECIMAL          DEFAULT 0,
    "is_paid"         BOOL             DEFAULT FALSE,
    "formula"         TEXT COLLATE "pg_catalog"."default",
    "note"            TEXT COLLATE "pg_catalog"."default",
    "paid_at"         TIMESTAMPTZ(6)
);
ALTER TABLE "employee_commissions" ADD CONSTRAINT "employee_commissions_employee_id_fkey" FOREIGN KEY ("employee_id") REFERENCES "employees" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

CREATE TABLE IF NOT EXISTS "employee_bonuses" (
    "id"          UUID PRIMARY KEY DEFAULT (uuid()),
    "employee_id" UUID NOT NULL,
    "amount"      INT8,
    "is_active"   BOOL             DEFAULT TRUE,
    "name"        TEXT COLLATE "pg_catalog"."default",
    "created_at"  TIMESTAMPTZ(6)     DEFAULT now()
);
ALTER TABLE "employee_bonuses"
    ADD CONSTRAINT "employee_bonuses_employee_id_fkey" FOREIGN KEY ("employee_id") REFERENCES "employees" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

CREATE TABLE IF NOT EXISTS "payrolls" (
    "id"                     uuid PRIMARY KEY DEFAULT (uuid()),
    "employee_id"            uuid NOT NULL,
    "total"                  int8 NOT NULL    DEFAULT 0,
    "month"                  int4,
    "year"                   int4,
    "commission_amount"      int8 NOT NULL    DEFAULT 0,
    "commission_explain"     json,
    "employee_rank_snapshot" json,
    "total_explain"          json,
    "project_bonus_amount"   int8 NOT NULL    DEFAULT 0,
    "due_date"               date,
    "project_bonus_explain"  json,
    "is_paid"                bool             DEFAULT FALSE,
    "conversion_amount"      int8 NOT NULL    DEFAULT 0,
    "base_salary_amount"     int8 NOT NULL    DEFAULT 0,
    "contract_amount"        int8
);

ALTER TABLE "payrolls" ADD CONSTRAINT "payrolls_employee_id_fkey" FOREIGN KEY ("employee_id") REFERENCES "employees" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

CREATE TABLE IF NOT EXISTS "project_commission_configs" (
    "id"              UUID PRIMARY KEY DEFAULT (uuid()),
    "deleted_at"      TIMESTAMP(6),
    "created_at"      TIMESTAMP(6)     DEFAULT (NOW()),
    "updated_at"      TIMESTAMP(6)     DEFAULT (NOW()),
    "project_id"      UUID,
    "position"        project_head_positions,
    "commission_rate" DECIMAL
);

ALTER TABLE "project_commission_configs" ADD CONSTRAINT "project_commission_configs_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "projects" ("id");

CREATE TABLE IF NOT EXISTS "cached_payrolls" (
    "id"       UUID PRIMARY KEY DEFAULT (uuid()),
    "month"    INT4,
    "year"     INT4,
    "batch"    INT4,
    "payrolls" JSON
);
ALTER TABLE "cached_payrolls" ADD CONSTRAINT "cached_payrolls_month_year_batch_key" UNIQUE ("month", "year", "batch");

-- +migrate Down
DROP TABLE IF EXISTS "base_salaries";
DROP TABLE IF EXISTS "accounting_categories";
DROP TABLE IF EXISTS "payrolls";
DROP TABLE IF EXISTS "employee_commissions";
DROP TABLE IF EXISTS "employee_bonuses";
DROP TABLE IF EXISTS "project_commission_configs";
DROP TABLE IF EXISTS "cached_payrolls";
DROP TABLE IF EXISTS "base_salaries";