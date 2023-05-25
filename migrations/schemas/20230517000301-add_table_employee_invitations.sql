-- +migrate Up
CREATE TABLE IF NOT EXISTS employee_invitations (
    id                          uuid PRIMARY KEY DEFAULT (uuid()),
    deleted_at                  TIMESTAMP(6),
    created_at                  TIMESTAMP(6)     DEFAULT (now()),
    updated_at                  TIMESTAMP(6)     DEFAULT (now()),

    employee_id                 uuid    NOT NULL,
    invited_by                  uuid    NOT NULL,
    invitation_code             TEXT    NOT NULL,
    is_completed                BOOLEAN NOT NULL DEFAULT false,
    is_info_updated             BOOLEAN NOT NULL DEFAULT false,
    is_discord_role_assigned    BOOLEAN NOT NULL DEFAULT false,
    is_basecamp_account_created BOOLEAN NOT NULL DEFAULT false,
    is_team_email_created       BOOLEAN NOT NULL DEFAULT false
);

ALTER TABLE employee_invitations
    ADD CONSTRAINT employee_invitations_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);
ALTER TABLE employee_invitations
    ADD CONSTRAINT employee_invitations_invited_by_fkey FOREIGN KEY (invited_by) REFERENCES employees (id);

-- +migrate Down
DROP TABLE IF EXISTS employee_invitations;
