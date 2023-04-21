-- +migrate Up
create table if not exists "audiences" (
    "id" uuid primary key default (uuid()),
    "created_at" timestamptz(6) default now(),
    "updated_at" timestamptz(6) default now(),
    "deleted_at" timestamptz(6),
    "email" text,
    "full_name" text,
    "first_name" text,
    "last_name" text,
    "source" text,
    "notion_id" text,
    "subscribed_dwarves_updates" boolean,
    "subscribed_techie_story" boolean,
    "subscribed_webuild" boolean
);

-- +migrate Down
drop table if exists "audiences";
