
-- +migrate Up
CREATE TABLE "countries" (
    "id" uuid PRIMARY KEY DEFAULT (uuid()),
    "deleted_at" TIMESTAMP(6),
    "created_at" TIMESTAMP(6) DEFAULT (now()),
    "updated_at" TIMESTAMP(6) DEFAULT (now()),
    "name" TEXT,
    "code" TEXT,
    "cities" jsonb default '[]'::jsonb
);

-- +migrate Down
DROP TABLE IF EXISTS "countries";
