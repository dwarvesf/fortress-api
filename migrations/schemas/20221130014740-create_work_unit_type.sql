
-- +migrate Up
CREATE TYPE work_unit_types AS ENUM (
  'development',
  'management',
  'training',
  'learning'
);

CREATE TYPE work_unit_statuses AS ENUM (
  'active',
  'archived'
);

ALTER TABLE work_units ALTER COLUMN status TYPE work_unit_statuses USING status::work_unit_statuses;

ALTER TABLE work_units ALTER COLUMN type TYPE work_unit_types USING type::work_unit_types;

-- +migrate Down
ALTER TABLE work_units ALTER COLUMN type TYPE TEXT USING type::TEXT;

ALTER TABLE work_units ALTER COLUMN status TYPE TEXT USING status::TEXT;

DROP TYPE IF EXISTS work_unit_types;

DROP TYPE IF EXISTS work_unit_statuses;
