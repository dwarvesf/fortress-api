-- +migrate Up

CREATE TABLE if not exists "accounting_transactions" (
    "id" uuid NOT NULL DEFAULT uuid(),
    "created_at" timestamp(8) DEFAULT now(),
    "deleted_at" timestamp(8),
    "date" timestamp(8) DEFAULT now(),
    "name" text,
    "amount" float8,
    "currency_id" uuid NOT NULL,
    "conversion_amount" int8,
    "organization" text,
    "metadata" json,
    "category" text,
    "currency" text,
    "conversion_rate" float4,
    "type" text,
    CONSTRAINT "accounting_transactions_currency_id_fkey" FOREIGN KEY ("currency_id") REFERENCES "currencies"("id") ON DELETE CASCADE,
    PRIMARY KEY ("id")
);

create table if not exists "assets" (
    "id" uuid primary key NOT NULL DEFAULT uuid(),
    "created_at" timestamp(8) DEFAULT now(),
    "deleted_at" timestamp(8),
    "name" text,
    "quantity" text,
    "price" int8,
    "currency_id" uuid,
    "location" text,
    "used_by" uuid,
    "purchased_at" date,
    "note" text,
    CONSTRAINT "assets_currency_id_fkey" FOREIGN KEY ("currency_id") REFERENCES "currencies"("id") ON DELETE CASCADE,
    CONSTRAINT "assets_used_by_fkey" FOREIGN KEY ("used_by") REFERENCES "employees"("id") ON DELETE CASCADE
);

create table if not exists "liabilities" (
    "id" uuid primary key NOT NULL DEFAULT uuid(),
    "created_at" timestamp(8) DEFAULT now(),
    "deleted_at" timestamp(8),
    "paid_at" timestamp(8),
    "name" text,
    "total" float8,
    "currency_id" uuid,
    CONSTRAINT "liabilities_currency_id_fkey" FOREIGN KEY ("currency_id") REFERENCES "currencies"("id") ON DELETE CASCADE
);

-- +migrate Down
drop table "accounting_transactions";
drop table "assets";
drop table "liabilities";
