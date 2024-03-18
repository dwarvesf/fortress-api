-- +migrate Up
ALTER TABLE "invoices" ADD COLUMN "bonus" DECIMAL;
ALTER TABLE "invoices" ADD COLUMN "total_without_bonus" DECIMAL;

-- +migrate Down
ALTER TABLE "invoices" DROP COLUMN "bonus";
ALTER TABLE "invoices" DROP COLUMN "total_without_bonus";
