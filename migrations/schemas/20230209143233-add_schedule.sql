-- +migrate Up
create table schedules (
    id uuid primary key default (uuid()),
    deleted_at timestamp(6),
    created_at timestamp(6) default (now()),
    updated_at timestamp(6) default (now()),
    synced_at timestamp(6) default (now()),

    name text,
    description text,
    start_time timestamp(6),
    end_time timestamp(6),
    schedule_type text
);

create table schedule_google_calendars (
    id uuid primary key default (uuid()),
    deleted_at timestamp(6),
    created_at timestamp(6) default (now()),
    updated_at timestamp(6) default (now()),

    schedule_id uuid references schedules(id),
    google_calendar_id text,
    description text,
    hangout_link text
);

create table schedule_discord_events (
    id uuid primary key default (uuid()),
    deleted_at timestamp(6),
    created_at timestamp(6) default (now()),
    updated_at timestamp(6) default (now()),

    schedule_id uuid references schedules(id),
    discord_event_id text,
    description text,
    voice_channel_id text
);

create table schedule_notion_pages (
    id uuid primary key default (uuid()),
    deleted_at timestamp(6),
    created_at timestamp(6) default (now()),
    updated_at timestamp(6) default (now()),

    schedule_id uuid references schedules(id),
    notion_page_id text,
    description text
);



-- +migrate Down
drop table schedule_discord_events;
drop table schedule_google_calendars;
drop table schedule_notion_pages;
drop table schedules;
