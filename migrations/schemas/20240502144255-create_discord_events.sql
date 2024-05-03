-- +migrate Up
CREATE TABLE discord_events (
    id uuid PRIMARY KEY DEFAULT (uuid()),
    title TEXT NOT NULL,
    description TEXT,
    date TIMESTAMP(6),
    event_url TEXT,
    msg_url TEXT,
    discord_event_type VARCHAR,
    created_at TIMESTAMP(6) DEFAULT (now()),
    updated_at TIMESTAMP(6) DEFAULT (now()),
    deleted_at TIMESTAMP(6) DEFAULT NULL
);

CREATE TABLE event_speakers (
    discord_event_id UUID NOT NULL,
    discord_account_id UUID NOT NULL,
    topic TEXT,
    UNIQUE (discord_event_id, discord_account_id)
);

ALTER TABLE event_speakers
    ADD CONSTRAINT event_speakers_discord_event_id_fkey FOREIGN KEY (discord_event_id) REFERENCES discord_events (id);

ALTER TABLE event_speakers
    ADD CONSTRAINT event_speakers_discord_account_id_fkey FOREIGN KEY (discord_account_id) REFERENCES discord_accounts (id);

-- +migrate Down

DROP TABLE event_speakers;
DROP TABLE discord_events;