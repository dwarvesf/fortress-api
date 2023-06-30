-- +migrate Up
CREATE TABLE IF NOT EXISTS "accounting_transactions" (
    "id"                UUID NOT NULL DEFAULT UUID(),
    "created_at"        TIMESTAMP(6)  DEFAULT NOW(),
    "deleted_at"        TIMESTAMP(6),

    "date"              TIMESTAMP(6)  DEFAULT NOW(),
    "name"              TEXT,
    "amount"            FLOAT8,
    "currency_id"       UUID NOT NULL,
    "conversion_amount" INT8,
    "organization"      TEXT,
    "metadata"          JSON,
    "category"          TEXT,
    "currency"          TEXT,
    "conversion_rate"   FLOAT8,
    "type"              TEXT,
    PRIMARY KEY ("id")
);

ALTER TABLE "accounting_transactions" ADD CONSTRAINT "accounting_transactions_currency_id_fkey" FOREIGN KEY ("currency_id") REFERENCES "currencies" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "accounting_transactions" ADD CONSTRAINT "transaction_info_unique" UNIQUE ("name", "date");

CREATE TABLE IF NOT EXISTS "assets" (
    "id"           UUID PRIMARY KEY NOT NULL DEFAULT UUID(),
    "created_at"   TIMESTAMP(6)              DEFAULT NOW(),
    "deleted_at"   TIMESTAMP(6),
    "name"         TEXT,
    "quantity"     TEXT,
    "price"        INT8,
    "currency_id"  UUID,
    "location"     TEXT,
    "used_by"      UUID,
    "purchased_at" date,
    "note"         TEXT,
    CONSTRAINT "assets_currency_id_fkey" FOREIGN KEY ("currency_id") REFERENCES "currencies" ("id") ON DELETE CASCADE,
    CONSTRAINT "assets_used_by_fkey" FOREIGN KEY ("used_by") REFERENCES "employees" ("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "liabilities" (
    "id"          UUID PRIMARY KEY NOT NULL DEFAULT UUID(),
    "created_at"  TIMESTAMP(6)              DEFAULT NOW(),
    "deleted_at"  TIMESTAMP(6),
    "paid_at"     TIMESTAMP(6),
    "name"        TEXT,
    "total"       FLOAT8,
    "currency_id" UUID,
    CONSTRAINT "liabilities_currency_id_fkey" FOREIGN KEY ("currency_id") REFERENCES "currencies" ("id") ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE "accounting_transactions";
DROP TABLE "assets";
DROP TABLE "liabilities";
