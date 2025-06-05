-- +migrate Up
CREATE TABLE
    IF NOT EXISTS inbound_fund_transactions (
        id uuid PRIMARY KEY DEFAULT (uuid ()),
        deleted_at TIMESTAMP(6),
        created_at TIMESTAMP(6) DEFAULT (now ()),
        updated_at TIMESTAMP(6) DEFAULT (now ()),
        invoice_id uuid NOT NULL,
        amount INT8 NOT NULL,
        notes TEXT,
        conversion_rate DECIMAL,
        paid_at TIMESTAMP(6),
        CONSTRAINT fk_invoices_inbound_fund_transactions FOREIGN KEY (invoice_id) REFERENCES invoices (id)
    );

-- +migrate Down
ALTER TABLE "invoices"
DROP COLUMN "inbound_fund_amount";

DROP TABLE IF EXISTS inbound_fund_transactions;
