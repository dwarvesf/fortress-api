
-- +migrate Up
CREATE TABLE IF NOT EXISTS project_member_positions (
    id                  uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at          TIMESTAMP(6),
    created_at          TIMESTAMP(6)     DEFAULT (now()),
    updated_at          TIMESTAMP(6)     DEFAULT (now()),
    project_member_id   uuid NOT NULL,
    position_id         uuid NOT NULL
);

ALTER TABLE project_member_positions
    ADD CONSTRAINT project_member_positions_project_member_id_fkey FOREIGN KEY (project_member_id) REFERENCES project_members(id);

ALTER TABLE project_member_positions
    ADD CONSTRAINT project_member_positions_position_id_fkey FOREIGN KEY (position_id) REFERENCES positions(id);

-- +migrate Down
ALTER TABLE project_member_positions DROP CONSTRAINT IF EXISTS project_member_positions_project_member_id_fkey;

ALTER TABLE project_member_positions DROP CONSTRAINT IF EXISTS project_member_positions_position_id_fkey;

DROP TABLE IF EXISTS project_member_positions;
