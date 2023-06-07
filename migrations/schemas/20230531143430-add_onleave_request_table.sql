-- +migrate Up
CREATE TABLE IF NOT EXISTS "on_leave_requests" (
    id                 UUID PRIMARY KEY DEFAULT (UUID()),
    deleted_at         TIMESTAMP(6),
    created_at         TIMESTAMP(6)     DEFAULT (now()),
    updated_at         TIMESTAMP(6)     DEFAULT (now()),

    title               TEXT NOT NULL,
    type                TEXT NOT NULL,
    start_date          DATE NOT NULL,
    end_date            DATE NOT NULL,
    shift               TEXT,
    description         TEXT,
    creator_id          UUID NOT NULL,
    approver_id         UUID NOT NULL,
    assignee_ids        JSONB
);

ALTER TABLE on_leave_requests
    ADD CONSTRAINT on_leave_requests_creator_id_fkey FOREIGN KEY (creator_id) REFERENCES employees (id);

ALTER TABLE on_leave_requests
    ADD CONSTRAINT on_leave_requests_approver_id_fkey FOREIGN KEY (approver_id) REFERENCES employees (id);

-- +migrate Down
DROP TABLE IF EXISTS on_leave_requests;
