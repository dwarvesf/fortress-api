-- +migrate Up
CREATE TABLE IF NOT EXISTS "expenses" (
    "id"                        UUID PRIMARY KEY                    DEFAULT (uuid()),
    "created_at"                TIMESTAMPTZ(6)                      DEFAULT NOW(),
    "updated_at"                TIMESTAMPTZ(6)                      DEFAULT NOW(),
    "deleted_at"                TIMESTAMPTZ(6),
    "employee_id"               UUID,
    "reason"                    TEXT COLLATE "pg_catalog"."default",
    "issued_date"               TIMESTAMPTZ(6)                      DEFAULT NOW(),
    "amount"                    INT8,
    "currency_id"               UUID,
    "currency"                  TEXT COLLATE "pg_catalog"."default" DEFAULT 'VND'::TEXT,
    "invoice_image_url"         TEXT COLLATE "pg_catalog"."default",
    "metadata"                  JSON,
    "accounting_transaction_id" UUID,
    "basecamp_id"               INT8
)
;
-- ----------------------------
-- Foreign Keys structure for table expense
-- ----------------------------
ALTER TABLE "expenses"
    ADD CONSTRAINT "expenses_accounting_transaction_id_fkey" FOREIGN KEY ("accounting_transaction_id") REFERENCES "accounting_transactions" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "expenses"
    ADD CONSTRAINT "expenses_currency_id_fkey" FOREIGN KEY ("currency_id") REFERENCES "currencies" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- +migrate Down
ALTER TABLE "expenses" DROP CONSTRAINT "expenses_accounting_transaction_id_fkey";
ALTER TABLE "expenses" DROP CONSTRAINT "expenses_currency_id_fkey";
DROP TABLE IF EXISTS "expenses";