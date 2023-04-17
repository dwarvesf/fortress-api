
-- +migrate Up
CREATE TABLE IF NOT EXISTS "expenses" (
  "id" uuid PRIMARY KEY DEFAULT (uuid()),
  "employee_id" uuid,
  "reason" text COLLATE "pg_catalog"."default",
  "issued_date" timestamptz(6) DEFAULT now(),
  "amount" int8,
  "created_at" timestamptz(6) DEFAULT now(),
  "updated_at" timestamptz(6) DEFAULT now(),
  "deleted_at" timestamptz(6),
  "currency_id" uuid,
  "currency" text COLLATE "pg_catalog"."default" DEFAULT 'VND'::text,
  "invoice_image_url" varchar(255) COLLATE "pg_catalog"."default",
  "metadata" json,
  "accounting_transaction_id" uuid,
  "basecamp_id" int8
)
;

-- ----------------------------
-- Foreign Keys structure for table expense
-- ----------------------------
ALTER TABLE "expenses" ADD CONSTRAINT "expenses_accounting_transaction_id_fkey" FOREIGN KEY ("accounting_transaction_id") REFERENCES "accounting_transactions" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;
ALTER TABLE "expenses" ADD CONSTRAINT "expenses_currency_id_fkey" FOREIGN KEY ("currency_id") REFERENCES "currencies" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;


-- +migrate Down
