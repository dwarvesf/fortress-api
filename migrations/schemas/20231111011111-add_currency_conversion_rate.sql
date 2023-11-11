-- +migrate Up
CREATE TABLE IF NOT EXISTS conversion_rates (
    id              UUID PRIMARY KEY DEFAULT (UUID()),
    deleted_at      TIMESTAMP(6),
    created_at      TIMESTAMP(6) DEFAULT (now()),
    updated_at      TIMESTAMP(6) DEFAULT (now()),

    currency_id     UUID DEFAULT NULL,
    to_usd          DECIMAL,
    to_vnd          DECIMAL
);

ALTER TABLE conversion_rates
    ADD CONSTRAINT conversion_rates_currency_id_fkey FOREIGN KEY (currency_id) REFERENCES currencies (id);

-- +migrate Down
DROP TABLE IF EXISTS conversion_rates;
