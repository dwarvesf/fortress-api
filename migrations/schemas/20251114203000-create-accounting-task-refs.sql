-- +migrate Up
CREATE TABLE IF NOT EXISTS accounting_task_refs (
    id UUID PRIMARY KEY DEFAULT uuid(),
    created_at TIMESTAMP(6) DEFAULT NOW(),
    updated_at TIMESTAMP(6),
    deleted_at TIMESTAMP(6),

    month INT NOT NULL,
    year INT NOT NULL,
    group_name TEXT NOT NULL,
    task_provider TEXT NOT NULL,
    task_ref TEXT NOT NULL,
    task_board TEXT,
    template_id UUID,
    project_id UUID,
    title TEXT,
    metadata JSON,

    UNIQUE (task_provider, task_ref)
);

-- +migrate Down
DROP TABLE IF EXISTS accounting_task_refs;
