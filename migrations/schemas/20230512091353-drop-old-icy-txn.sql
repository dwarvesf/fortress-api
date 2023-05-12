-- +migrate Up

DROP VIEW "vw_icy_treasury_funds";

DROP TABLE "icy_transactions_out";

DROP TABLE "icy_transactions_in";

DROP TABLE "icy_treasury_categories";

-- +migrate Down

CREATE TABLE "icy_treasury_categories" (
  "id" uuid NOT NULL DEFAULT uuid(),
  "created_at" timestamp(8) DEFAULT NOW(),
  "deleted_at" timestamp(8),
  "name" text,
  "category_manager_id" uuid,
  CONSTRAINT "icy_treasury_categories_category_manager_id_fkey" FOREIGN KEY ("category_manager_id") REFERENCES "employees" ("id"),
  PRIMARY KEY ("id")
);

CREATE TABLE "icy_transactions_out" (
  "id" uuid NOT NULL DEFAULT uuid(),
  "created_at" timestamp(8) DEFAULT NOW(),
  "deleted_at" timestamp(8),
  "amount" text,
  "description" text,
  "category_id" uuid,
  "to_employee_id" uuid NOT NULL,
  "approver_id" uuid,
  CONSTRAINT "icy_transactions_category_id_fkey" FOREIGN KEY ("category_id") REFERENCES "icy_treasury_categories" ("id"),
  CONSTRAINT "icy_transactions_to_employee_id_fkey" FOREIGN KEY ("to_employee_id") REFERENCES "employees" ("id"),
  CONSTRAINT "icy_transactions_approver_id_fkey" FOREIGN KEY ("approver_id") REFERENCES "employees" ("id"),
  PRIMARY KEY ("id")
);

CREATE TABLE "icy_transactions_in" (
  "id" uuid NOT NULL DEFAULT uuid(),
  "created_at" timestamp(8) DEFAULT NOW(),
  "deleted_at" timestamp(8),
  "date" timestamp(8) DEFAULT NOW(),
  "description" text,
  "amount" text,
  "category_id" uuid,
  CONSTRAINT "icy_transactions_category_id_fkey" FOREIGN KEY ("category_id") REFERENCES "icy_treasury_categories" ("id"),
  PRIMARY KEY ("id")
);

CREATE VIEW "vw_icy_treasury_funds" AS
SELECT
  t1.category_id,
  t2.total_in - t1.total_out AS balance
FROM
  (
    SELECT
      category_id,
      SUM(amount :: NUMERIC) AS total_out
    FROM
      icy_transactions_out ito
    GROUP BY
      category_id
  ) t1
  JOIN (
    SELECT
      category_id,
      SUM(amount :: NUMERIC) AS total_in
    FROM
      icy_transactions_in iti
    GROUP BY
      category_id
  ) t2 ON t1.category_id = t2.category_id;
