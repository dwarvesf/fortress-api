
-- +migrate Up
CREATE TABLE IF NOT EXISTS public.operational_services
(
    id            uuid not null primary key,
    name          varchar(255),
    amount        bigint,
    currency_id   uuid references public.currencies(id),
    note          varchar(255),
    register_date date,
    start_at      date      default now(),
    end_at        date,
    is_active     boolean   default true,
    created_at    timestamp default now(),
    update_at     timestamp default now(),
    deleted_at    timestamp
    );
-- +migrate Down
DROP TABLE public.operational_services;
