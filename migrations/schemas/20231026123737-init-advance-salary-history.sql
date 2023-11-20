-- +migrate Up
CREATE TABLE IF NOT EXISTS salary_advance_histories (
    id          uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at  TIMESTAMP(6),
    created_at  TIMESTAMP(6)     DEFAULT (now()),
    updated_at  TIMESTAMP(6)     DEFAULT (now()),
    amount_icy        int8 NOT NULL DEFAULT 0,
    amount_usd        DECIMAL NOT NULL DEFAULT 0,
    base_amount       DECIMAL NOT NULL DEFAULT 0,
    is_paid_back      BOOLEAN   DEFAULT FALSE,
    paid_at           TIMESTAMP(6),
    conversion_rate   DECIMAL          DEFAULT 0,
    currency_id       uuid,
    employee_id       uuid
);

ALTER TABLE salary_advance_histories ADD CONSTRAINT salary_advance_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);
ALTER TABLE salary_advance_histories ADD CONSTRAINT salary_advance_currency_id_fkey FOREIGN KEY (currency_id) REFERENCES currencies (id);

-- +migrate Down
ALTER TABLE salary_advance_histories DROP CONSTRAINT salary_advance_employee_id_fkey;
ALTER TABLE salary_advance_histories DROP CONSTRAINT salary_advance_currency_id_fkey;
DROP TABLE IF EXISTS salary_advance_histories;
