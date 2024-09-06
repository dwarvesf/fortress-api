
-- +migrate Up
ALTER TABLE event_speakers
    DROP CONSTRAINT IF EXISTS event_speakers_event_id_discord_account_id_key;

ALTER TABLE event_speakers
    ADD CONSTRAINT event_speakers_topic_key UNIQUE (topic);



-- +migrate Down
ALTER TABLE event_speakers
    DROP CONSTRAINT IF EXISTS event_speakers_topic_key;

ALTER TABLE event_speakers ADD CONSTRAINT event_speakers_event_id_discord_account_id_key UNIQUE (event_id, discord_account_id);
