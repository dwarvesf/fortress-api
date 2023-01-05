
-- +migrate Up
CREATE TYPE project_functions AS ENUM (
    'development',
    'learning',
    'training',
    'management'
);
ALTER TABLE "projects" ADD COLUMN "function" project_functions;

-- +migrate Down
ALTER TABLE "projects" DROP COLUMN "function";
DROP TYPE project_functions;
