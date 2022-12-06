-- +migrate Up
CREATE TYPE "event_types" AS ENUM (
  'feedback',
  'survey'
);

CREATE TYPE "event_subtypes" AS ENUM (
  'peer-review',
  'engagement',
  'work',
  'appreciation',
  'comment'
);

CREATE TYPE "relationships" AS ENUM (
  'peer',
  'line-manager',
  'chapter-lead',
  'self'
);

CREATE TABLE IF NOT EXISTS "feedback_events" (
  "id" uuid PRIMARY KEY DEFAULT uuid(),
  "deleted_at" TIMESTAMP(6),
  "created_at" TIMESTAMP(6) DEFAULT NOW(),
  "updated_at" TIMESTAMP(6) DEFAULT NOW(),

  "title" TEXT,
  "type" event_types,
  "subtype" event_subtypes,
  "status" TEXT,
  "created_by" uuid NOT NULL,
  "start_date" TIMESTAMP(6),
  "end_date" TIMESTAMP(6)
);

CREATE TABLE IF NOT EXISTS "employee_event_topics" (
  "id" uuid PRIMARY KEY DEFAULT uuid(),
  "deleted_at" TIMESTAMP(6),
  "created_at" TIMESTAMP(6) DEFAULT NOW(),
  "updated_at" TIMESTAMP(6) DEFAULT NOW(),

  "title" TEXT,
  "event_id" uuid NOT NULL,
  "employee_id" uuid NOT NULL,
  "project_id" uuid NULL
);

CREATE TABLE IF NOT EXISTS "employee_event_reviewers" (
  "id" uuid PRIMARY KEY DEFAULT uuid(),
  "deleted_at" TIMESTAMP(6),
  "created_at" TIMESTAMP(6) DEFAULT NOW(),
  "updated_at" TIMESTAMP(6) DEFAULT NOW(),

  "employee_event_topic_id" uuid NOT NULL,
  "reviewer_id" uuid NOT NULL,
  "status" TEXT,
  "relationship" relationships,
  "is_shared" BOOL DEFAULT FALSE,
  "is_read" BOOL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS "employee_event_questions" (
  "id" uuid PRIMARY KEY DEFAULT uuid(),
  "deleted_at" TIMESTAMP(6),
  "created_at" TIMESTAMP(6) DEFAULT NOW(),
  "updated_at" TIMESTAMP(6) DEFAULT NOW(),

  "employee_event_reviewer_id" uuid NOT NULL,
  "content" TEXT,
  "answers" TEXT,
  "note" TEXT
);

CREATE TABLE IF NOT EXISTS "questions" (
  "id" uuid PRIMARY KEY DEFAULT uuid(),
  "deleted_at" TIMESTAMP(6),
  "created_at" TIMESTAMP(6) DEFAULT NOW(),
  "updated_at" TIMESTAMP(6) DEFAULT NOW(),

  "type" TEXT,
  "subtype" TEXT,
  "content" TEXT
);

ALTER TABLE "employee_event_topics" ADD CONSTRAINT employee_event_topics_event_id_fkey  FOREIGN KEY ("event_id") REFERENCES "feedback_events" ("id");

ALTER TABLE "employee_event_topics" ADD CONSTRAINT employee_event_topics_project_id_fkey  FOREIGN KEY ("project_id") REFERENCES "projects" ("id");

ALTER TABLE "employee_event_questions" ADD CONSTRAINT employee_event_questions_employee_event_reviewer_id_fkey FOREIGN KEY ("employee_event_reviewer_id") REFERENCES "employee_event_reviewers" ("id");

ALTER TABLE "employee_event_reviewers" ADD CONSTRAINT employee_event_reviewers_employee_event_topic_id_fkey FOREIGN KEY ("employee_event_topic_id") REFERENCES "employee_event_topics" ("id");

ALTER TABLE "employee_event_reviewers" ADD CONSTRAINT employee_event_reviewers_reviewer_id_fkey FOREIGN KEY ("reviewer_id") REFERENCES "employees" ("id");

ALTER TABLE "employee_event_topics" ADD CONSTRAINT employee_event_topics_employee_id_fkey FOREIGN KEY ("employee_id") REFERENCES "employees" ("id");

ALTER TABLE "feedback_events" ADD CONSTRAINT feedback_events_created_by_fkey FOREIGN KEY ("created_by") REFERENCES "employees" ("id");

-- +migrate Down
ALTER TABLE "employee_event_topics" DROP CONSTRAINT IF EXISTS employee_event_topics_event_id_fkey;
ALTER TABLE "employee_event_topics" DROP CONSTRAINT IF EXISTS employee_event_topics_project_id_fkey;
ALTER TABLE "employee_event_questions" DROP CONSTRAINT IF EXISTS employee_event_questions_employee_event_reviewer_id_fkey;
ALTER TABLE "employee_event_reviewers" DROP CONSTRAINT IF EXISTS employee_event_reviewers_employee_event_topic_id_fkey;
ALTER TABLE "employee_event_reviewers" DROP CONSTRAINT IF EXISTS employee_event_reviewers_reviewer_id_fkey;
ALTER TABLE "employee_event_topics" DROP CONSTRAINT IF EXISTS employee_event_topics_employee_id_fkey;
ALTER TABLE "feedback_events" DROP CONSTRAINT IF EXISTS feedback_events_created_by_fkey;

DROP TABLE IF EXISTS "questions";
DROP TABLE IF EXISTS "employee_event_questions";
DROP TABLE IF EXISTS "employee_event_reviewers";
DROP TABLE IF EXISTS "employee_event_topics";
DROP TABLE IF EXISTS "feedback_events";

DROP TYPE IF EXISTS event_types;
DROP TYPE IF EXISTS event_subtypes;
DROP TYPE IF EXISTS relationships;
