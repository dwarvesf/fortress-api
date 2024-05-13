
-- +migrate Up
CREATE TABLE IF NOT EXISTS memo_authors  (
  memo_log_id UUID NOT NULL REFERENCES memo_logs(id),
  community_member_id UUID NOT NULL REFERENCES community_members(id),
  created_at    TIMESTAMP(6)     DEFAULT (now()),
  PRIMARY KEY (memo_log_id, community_member_id)
);

ALTER TABLE memo_logs ADD UNIQUE ("url");
-- +migrate Down
ALTER TABLE memo_logs DROP CONSTRAINT memo_logs_url_key;

DROP TABLE IF EXISTS memo_authors;
