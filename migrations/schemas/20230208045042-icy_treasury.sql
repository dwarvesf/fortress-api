-- +migrate Up
create table "icy_treasury_categories" (
  "id" uuid not null default uuid(),
  "created_at" timestamp(8) DEFAULT now(),
  "deleted_at" timestamp(8),
  "name" text,
  "category_manager_id" uuid,
  constraint "icy_treasury_categories_category_manager_id_fkey" foreign key ("category_manager_id") references "employees"("id"),
  primary key ("id")
);

Create table "icy_transactions_out" (
  "id" uuid NOT NULL DEFAULT uuid(),
  "created_at" timestamp(8) DEFAULT now(),
  "deleted_at" timestamp(8),
  "amount" text,
  "description" text,
  "category_id" uuid,
  "to_employee_id" uuid not null,
  "approver_id" uuid,
  constraint "icy_transactions_category_id_fkey" foreign key ("category_id") references "icy_treasury_categories"("id"),
  constraint "icy_transactions_to_employee_id_fkey" foreign key ("to_employee_id") references "employees"("id"),
  constraint "icy_transactions_approver_id_fkey" foreign key ("approver_id") references "employees"("id"),
  primary key ("id")
);

Create table "icy_transactions_in" (
  "id" uuid NOT NULL DEFAULT uuid(),
  "created_at" timestamp(8) DEFAULT now(),
  "deleted_at" timestamp(8),
  "date" timestamp(8) DEFAULT now(),
  "description" text,
  "amount" text,
  "category_id" uuid,
  constraint "icy_transactions_category_id_fkey" foreign key ("category_id") references "icy_treasury_categories"("id"),
  primary key ("id")
);

create view "vw_icy_treasury_funds" as
SELECT
	t1.category_id,
	t2.total_in - t1.total_out as balance
FROM (
	SELECT
		category_id,
		sum(amount::NUMERIC) AS total_out
	FROM
		icy_transactions_out ito
	GROUP BY
		category_id) t1
	JOIN (
		SELECT
			category_id,
			sum(amount::NUMERIC) AS total_in
		FROM
			icy_transactions_in iti
		GROUP BY
			category_id) t2 ON t1.category_id = t2.category_id;


-- +migrate Down
drop view "vw_icy_treasury_funds",
drop table "icy_transactions_out",
drop table "icy_transactions_in",
drop table "icy_treasury_categories";
