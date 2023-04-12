
-- +migrate Up
CREATE TABLE IF NOT EXISTS public.operational_services (
    id            uuid PRIMARY KEY DEFAULT (uuid()),
    created_at    TIMESTAMP(6) DEFAULT NOW(),
    update_at     TIMESTAMP(6) DEFAULT NOW(),
    deleted_at    TIMESTAMP(6),
    name          TEXT,
    amount        INT8,
    currency_id   UUID REFERENCES public.currencies (id),
    note          TEXT,
    register_date DATE      DEFAULT NOW(),
    start_at      DATE      DEFAULT NOW(),
    end_at        DATE,
    is_active     BOOLEAN   DEFAULT TRUE
);
-- +migrate Down
DROP TABLE public.operational_services;
