-- +migrate Up
DROP TABLE schedule_discord_events;
DROP TABLE schedule_google_calendars;
DROP TABLE schedule_notion_pages;
DROP TABLE schedules;

-- +migrate Down
CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT (UUID()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6) DEFAULT (NOW()),
    updated_at TIMESTAMP(6) DEFAULT (NOW()),
    synced_at TIMESTAMP(6) DEFAULT (NOW()),

    name TEXT,
    description TEXT,
    start_time TIMESTAMP(6),
    end_time TIMESTAMP(6),
    schedule_type TEXT
);

CREATE TABLE schedule_google_calendars (
    id UUID PRIMARY KEY DEFAULT (UUID()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6) DEFAULT (NOW()),
    updated_at TIMESTAMP(6) DEFAULT (NOW()),

    schedule_id UUID REFERENCES schedules(id),
    google_calendar_id TEXT,
    description TEXT,
    hangout_link TEXT
);

CREATE TABLE schedule_discord_events (
    id UUID PRIMARY KEY DEFAULT (UUID()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6) DEFAULT (NOW()),
    updated_at TIMESTAMP(6) DEFAULT (NOW()),

    schedule_id UUID REFERENCES schedules(id),
    discord_event_id TEXT,
    description TEXT,
    voice_channel_id TEXT
);

CREATE TABLE schedule_notion_pages (
    id UUID PRIMARY KEY DEFAULT (UUID()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6) DEFAULT (NOW()),
    updated_at TIMESTAMP(6) DEFAULT (NOW()),

    schedule_id UUID REFERENCES schedules(id),
    notion_page_id TEXT,
    description TEXT
);




