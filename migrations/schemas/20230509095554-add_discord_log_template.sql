-- +migrate Up
create table if not EXISTS discord_log_templates (
    id uuid PRIMARY KEY DEFAULT(uuid()),
    deleted_at TIMESTAMP(6),
    created_at TIMESTAMP(6) DEFAULT(NOW()),
    updated_at TIMESTAMP(6) DEFAULT(NOW()),

    type TEXT,
    content TEXT
);

-- +migrate Down
drop table discord_log_templates;
