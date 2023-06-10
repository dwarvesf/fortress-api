-- +migrate Up
CREATE TYPE project_important_level_type AS ENUM (
    'low',
    'medium-',
    'medium',
    'medium+',
    'high'
);

ALTER TABLE projects ADD COLUMN account_rating INT NOT NULL DEFAULT 0;
ALTER TABLE projects ADD COLUMN delivery_rating INT NOT NULL  DEFAULT 0;
ALTER TABLE projects ADD COLUMN lead_rating INT NOT NULL  DEFAULT 0;
ALTER TABLE projects ADD COLUMN important_level project_important_level_type DEFAULT NULL;

-- +migrate Down
ALTER TABLE projects DROP COLUMN account_rating;
ALTER TABLE projects DROP COLUMN delivery_rating;
ALTER TABLE projects DROP COLUMN lead_rating;
ALTER TABLE projects DROP COLUMN important_level;
DROP TYPE project_important_level_type;
