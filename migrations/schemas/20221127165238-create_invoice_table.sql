
-- +migrate Up
CREATE TABLE IF NOT EXISTS currencies (
    id uuid     PRIMARY KEY  DEFAULT (uuid()),
    deleted_at  TIMESTAMP(6),
    created_at  TIMESTAMP(6) DEFAULT (now()),
    updated_at  TIMESTAMP(6) DEFAULT (now()),
    name        TEXT,
    symbol      TEXT,
    locale      varchar(6),
    type        TEXT         DEFAULT 'fiat'::TEXT
);

CREATE TABLE IF NOT EXISTS bank_accounts (
    id uuid         PRIMARY KEY  DEFAULT (uuid()),
    deleted_at      TIMESTAMP(6),
    created_at      TIMESTAMP(6) DEFAULT (now()),
    updated_at  TIMESTAMP(6) DEFAULT (now()),
    account_number  TEXT NOT NULL,
    bank_name       TEXT,
    currency_id     uuid NOT NULL,
    owner_name      TEXT,
    address         TEXT,
    swift_code      TEXT,
    routing_number  varchar(255),
    name            varchar(255),
    uk_sort_code    varchar(12)
);

ALTER TABLE bank_accounts
    ADD CONSTRAINT bank_accounts_currency_id_fkey FOREIGN KEY (currency_id) REFERENCES currencies(id);


CREATE TYPE  invoice_statuses AS ENUM (
  'draft',
  'sent',
  'overdue',
  'paid',
  'error',
  'scheduled'
);

CREATE TABLE invoices (
    id                uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at        TIMESTAMP(6),
    created_at        TIMESTAMP(6)     DEFAULT (now()),
    updated_at        TIMESTAMP(6)     DEFAULT (now()),
    number            TEXT,
    due_at            TIMESTAMP(6),
    project_id        uuid NOT NULL,
    description       TEXT,
    bank_id           uuid NOT NULL,
    sub_total         int4,
    tax               int4,
    discount          int4,
    invoice_file_url  TEXT,
    error_invoice_id  uuid NOT NULL,
    metadata          json,
    paid_at           TIMESTAMP(6),
    note              TEXT,
    line_items        json,
    failed_at         TIMESTAMP(6),
    month             int4,
    year              int4,
    invoiced_at       date,
    status            invoice_statuses,
    email             varchar(255),
    cc                json,
    thread_id         varchar(255),
    sent_by           uuid NOT NULL,
    total             numeric,
    conversion_amount int8,
    scheduled_date    TIMESTAMP(6),
    conversion_rate   float4
);

ALTER TABLE invoices
    ADD CONSTRAINT invoices_sent_by_fkey FOREIGN KEY (sent_by) REFERENCES employees(id);

ALTER TABLE invoices
    ADD CONSTRAINT invoices_err_invoice FOREIGN KEY (error_invoice_id) REFERENCES invoices(id);

ALTER TABLE invoices
    ADD CONSTRAINT invoices_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id);

ALTER TABLE invoices
    ADD CONSTRAINT invoices_bank_id_fkey FOREIGN KEY (bank_id) REFERENCES bank_accounts(id);

-- +migrate Down
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS bank_accounts;
DROP TABLE IF EXISTS currencies;
DROP TYPE IF EXISTS invoice_statuses;
