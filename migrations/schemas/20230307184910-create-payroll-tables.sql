-- +migrate Up
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
    "updated_at" TIMESTAMPTZ(6) DEFAULT NOW(),
    "deleted_at" timestamptz(6),
    "is_active" bool DEFAULT true,
    "batch" int4,
    "type" text COLLATE "pg_catalog"."default",
    "category" text COLLATE "pg_catalog"."default"
);

ALTER TABLE
    "base_salaries"
ADD
    CONSTRAINT "base_salaries_currency_id_fkey" FOREIGN KEY ("currency_id") REFERENCES "currencies" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

ALTER TABLE
    "base_salaries"
ADD
    CONSTRAINT "base_salaries_employee_id_fkey" FOREIGN KEY ("employee_id") REFERENCES "employees" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

CREATE TABLE IF NOT EXISTS "accounting_transactions" (
    "id" uuid PRIMARY KEY DEFAULT (uuid()),
    "created_at" timestamptz(6) DEFAULT now(),
    "updated_at" TIMESTAMPTZ(6) DEFAULT NOW(),
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
);

ALTER TABLE
    "accounting_transactions"
ADD
    CONSTRAINT "transaction_info_unique" UNIQUE ("name", "date");

ALTER TABLE
    "accounting_transactions"
ADD
    CONSTRAINT "accounting_transactions_pkey" PRIMARY KEY ("id");

ALTER TABLE
    "accounting_transactions"
ADD
    CONSTRAINT "accounting_transactions_currency_id_fkey" FOREIGN KEY ("currency_id") REFERENCES "currencies" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

CREATE TABLE IF NOT EXISTS "accounting_categories" (
    "id" uuid PRIMARY KEY DEFAULT (uuid()),
    "created_at" timestamptz(6) DEFAULT now(),
    "updated_at" TIMESTAMPTZ(6) DEFAULT NOW(),
    "deleted_at" timestamptz(6),
    "name" text COLLATE "pg_catalog"."default",
    "type" text COLLATE "pg_catalog"."default"
);

BEGIN;

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        '83eb7df1-7c37-4d20-aeab-10c86499e7ae',
        '2020-04-16 12:01:54.30488+00',
        NULL,
        'Payroll',
        'SE'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        '42e22452-9abd-48c2-800e-884a4d451083',
        '2019-02-20 04:24:14.967209+00',
        NULL,
        'Payroll for Operation',
        'OV'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        '93736ed2-4c86-4d0d-bcd2-a9be7d71554b',
        '2020-06-02 04:25:34.048356+00',
        NULL,
        'Commission',
        'SE'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        '18dc5d8e-fc8e-4db2-8eac-38bd956531d9',
        '2019-10-23 08:35:51.109698+00',
        NULL,
        'In',
        'In'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        'e9384e3e-fad8-439c-8e9e-4f5f2647f13c',
        '2019-02-20 04:24:14.967209+00',
        NULL,
        'Payroll for Design',
        'SE'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        'a3e29d7c-8ec8-4af8-9384-0c04810ae379',
        '2019-02-20 04:24:14.967209+00',
        NULL,
        'Payroll for Engineer',
        'SE'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        'd79d9cff-d564-4d15-9ded-223dee7839e4',
        '2019-02-20 04:24:14.967209+00',
        NULL,
        'Payroll for Sales',
        'SE'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        'ad98f449-8027-4387-9a78-f4c8afbcde1c',
        '2019-09-18 08:05:50.015091+00',
        NULL,
        'Payroll for Venture',
        'SE'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        '3f345cb6-cc23-416b-991e-268aa17a6a8f',
        '2019-02-20 04:24:14.967209+00',
        NULL,
        'Server/Hosting',
        'OP'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        'afcf2089-73b9-45e8-ab12-f8d6c430ede3',
        '2019-02-20 04:24:14.967209+00',
        NULL,
        'Tools',
        'OV'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        'c529e27e-b419-47e0-88ab-d1717b65ecce',
        '2019-09-18 08:09:54.676547+00',
        NULL,
        'Payroll for Middle Mngr',
        'OV'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        'ab2e079e-83f2-4243-906e-c7a00d4f4594',
        '2019-02-20 04:24:14.967209+00',
        NULL,
        'Payroll for Marketing',
        'OV'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        '734895da-97d9-4be3-bed0-678468cf4c7b',
        '2019-02-20 04:24:14.967209+00',
        NULL,
        'Payroll for Recruit',
        'OV'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        '5b1e1334-47a2-4619-b0b5-5e1f768a85a3',
        '2019-09-18 08:10:47.353348+00',
        NULL,
        'Vietnam HR-legal service',
        'OV'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        '2862c715-fbd0-4b00-aacb-b92a4254215d',
        '2019-02-20 04:24:14.967209+00',
        NULL,
        'Office Supply',
        'OV'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        'eb68043d-beeb-4eed-a8a8-3db2077aeb7c',
        '2019-02-20 04:24:14.967209+00',
        NULL,
        'Company Trip',
        'CA'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        'e5612457-6bed-4391-9132-cb40493daeb0',
        '2019-02-20 04:24:14.967209+00',
        NULL,
        'Training',
        'CA'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        '065f67ef-9486-422c-a328-70fddfc38adf',
        '2019-02-20 04:24:14.967209+00',
        NULL,
        'Assets',
        'CA'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        '16a9728f-83e9-453c-91dd-61afb468cc86',
        '2019-02-20 04:24:14.967209+00',
        NULL,
        'Insurance',
        'CA'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        '07c8045b-8008-4105-af3b-9ccf2d5c9ca9',
        '2019-02-20 04:24:14.967209+00',
        NULL,
        'Office Services',
        'OP'
    );

INSERT INTO
    "accounting_categories" ("created_at", "deleted_at", "name", "type")
VALUES
    (
        'edcb58ba-85c9-43d6-b99b-52cbcda1595e',
        '2019-02-20 04:24:14.967209+00',
        NULL,
        'Office Space',
        'OP'
    );

COMMIT;

ALTER TABLE
    "accounting_categories"
ADD
    CONSTRAINT "accounting_categories_pkey" PRIMARY KEY ("id");

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

CREATE TABLE IF NOT EXISTS "employee_bonuses" (
    "id" uuid PRIMARY KEY DEFAULT (uuid()),
    "employee_id" uuid NOT NULL,
    "amount" int8,
    "is_active" bool DEFAULT true,
    "name" text COLLATE "pg_catalog"."default",
    "created_at" timestamp(6) DEFAULT now(),
    "updated_at" TIMESTAMPTZ(6) DEFAULT NOW()
);

ALTER TABLE
    "employee_bonuses"
ADD
    CONSTRAINT "employee_bonuses_employee_id_fkey" FOREIGN KEY ("employee_id") REFERENCES "employees" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

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

CREATE TABLE IF NOT EXISTS project_commission_configs (
    id UUID PRIMARY KEY DEFAULT (uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6) DEFAULT (NOW()),
    updated_at TIMESTAMP(6) DEFAULT (NOW()),
    project_id UUID,
    position project_head_positions,
    commission_rate DECIMAL
);

ALTER TABLE
    project_commission_configs
ADD
    CONSTRAINT project_commission_configs_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects (id);

-- +migrate Down