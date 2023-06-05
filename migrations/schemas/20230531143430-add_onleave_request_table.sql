-- +migrate Up
CREATE TABLE IF NOT EXISTS "on_leave_requests" (
    id                 UUID PRIMARY KEY DEFAULT (UUID()),
    deleted_at         TIMESTAMP(6),
    created_at         TIMESTAMP(6)     DEFAULT (now()),
    updated_at         TIMESTAMP(6)     DEFAULT (now()),

    name                TEXT NOT NULL,
    off_type            TEXT NOT NULL,
    start_date          DATE,
    end_date            DATE,
    shift               TEXT,
    title               TEXT NOT NULL,
    description         TEXT,
    creator_id          UUID NOT NULL,
    approver_id         UUID NOT NULL,
    assignee_ids        TEXT[]
);

ALTER TABLE on_leave_request
    ADD CONSTRAINT on_leave_request_creator_id_fkey FOREIGN KEY (creator_id) REFERENCES employees (id);

ALTER TABLE on_leave_request
    ADD CONSTRAINT on_leave_request_approver_id_fkey FOREIGN KEY (approver_id) REFERENCES employees (id);

-- +migrate Down
DROP TABLE IF EXISTS on_leave_requests;
