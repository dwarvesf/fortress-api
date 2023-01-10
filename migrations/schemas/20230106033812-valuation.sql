-- +migrate Up
create or replace view vw_account_receivables as
 SELECT COALESCE(vnd.vnd, 0::numeric) AS vnd,
    COALESCE(usd.usd, 0::numeric) AS usd,
    COALESCE(sgd.sgd, 0::numeric) AS sgd,
    COALESCE(eur.eur, 0::numeric) AS eur,
    COALESCE(gbp.gbp, 0::numeric) AS gbp,
    timetable.year
   FROM ( SELECT invoices.year
           FROM invoices
          GROUP BY invoices.year) timetable
     LEFT JOIN ( SELECT COALESCE(sum(invoices.total), 0::numeric) AS vnd,
            invoices.year
           FROM invoices
          WHERE (invoices.status not in ('draft', 'paid', 'error')) AND (invoices.bank_id IN ( SELECT bank_accounts.id
                   FROM bank_accounts
                  WHERE (bank_accounts.currency_id IN ( SELECT currencies.id
                           FROM currencies
                          WHERE currencies.name = 'VND'::text))))
          GROUP BY invoices.year) vnd ON timetable.year = vnd.year
     LEFT JOIN ( SELECT COALESCE(sum(invoices.total), 0::numeric) AS usd,
            invoices.year
           FROM invoices
          WHERE (invoices.status not in ('draft', 'paid', 'error')) AND (invoices.bank_id IN ( SELECT bank_accounts.id
                   FROM bank_accounts
                  WHERE (bank_accounts.currency_id IN ( SELECT currencies.id
                           FROM currencies
                          WHERE currencies.name = 'USD'::text))))
          GROUP BY invoices.year) usd ON timetable.year = usd.year
     LEFT JOIN ( SELECT COALESCE(sum(invoices.total), 0::numeric) AS sgd,
            invoices.year
           FROM invoices
          WHERE (invoices.status not in ('draft', 'paid', 'error')) AND (invoices.bank_id IN ( SELECT bank_accounts.id
                   FROM bank_accounts
                  WHERE (bank_accounts.currency_id IN ( SELECT currencies.id
                           FROM currencies
                          WHERE currencies.name = 'SGD'::text))))
          GROUP BY invoices.year) sgd ON timetable.year = usd.year
     LEFT JOIN ( SELECT COALESCE(sum(invoices.total), 0::numeric) AS eur,
            invoices.year
           FROM invoices
          WHERE (invoices.status not in ('draft', 'paid', 'error')) AND (invoices.bank_id IN ( SELECT bank_accounts.id
                   FROM bank_accounts
                  WHERE (bank_accounts.currency_id IN ( SELECT currencies.id
                           FROM currencies
                          WHERE currencies.name = 'EUR'::text))))
          GROUP BY invoices.year) eur ON timetable.year = eur.year
     LEFT JOIN ( SELECT COALESCE(sum(invoices.total), 0::numeric) AS gbp,
            invoices.year
           FROM invoices
          WHERE (invoices.status not in ('draft', 'paid', 'error')) AND (invoices.bank_id IN ( SELECT bank_accounts.id
                   FROM bank_accounts
                  WHERE (bank_accounts.currency_id IN ( SELECT currencies.id
                           FROM currencies
                          WHERE currencies.name = 'GBP'::text))))
          GROUP BY invoices.year) gbp ON timetable.year = gbp.year
  ORDER BY timetable.year DESC;

create or replace view vw_liabilities as
SELECT
	COALESCE(vnd.vnd, 0::numeric) AS vnd,
	COALESCE(usd.usd, 0::numeric) AS usd,
	COALESCE(sgd.sgd, 0::numeric) AS sgd,
	COALESCE(eur.eur, 0::numeric) AS eur,
	COALESCE(gbp.gbp, 0::numeric) AS gbp,
	timetable.year
FROM (
	SELECT
		date_part('year', liabilities.created_at) AS year
	FROM
		liabilities
	GROUP BY
		year) timetable
	LEFT JOIN (
		SELECT
			COALESCE(sum(liabilities.total), 0::numeric) AS vnd,
			date_part('year', liabilities.created_at) AS year
		FROM
			liabilities
		WHERE (liabilities.paid_at IS NULL)
		AND(liabilities.currency_id IN(
				SELECT
					id FROM currencies
				WHERE
					name = 'VND'::text))
	GROUP BY
		year) vnd ON timetable.year = vnd.year
	LEFT JOIN (
		SELECT
			COALESCE(sum(liabilities.total), 0::numeric) AS usd, date_part('year', liabilities.created_at) AS year
		FROM
			liabilities
		WHERE (liabilities.paid_at IS NULL)
		AND(liabilities.currency_id IN(
				SELECT
					id FROM currencies
				WHERE
					name = 'USD'::text))
	GROUP BY
		year) usd ON timetable.year = usd.year
	LEFT JOIN (
		SELECT
			COALESCE(sum(liabilities.total), 0::numeric) AS eur, date_part('year', liabilities.created_at) AS year
		FROM
			liabilities
		WHERE (liabilities.paid_at IS NULL)
		AND(liabilities.currency_id IN(
				SELECT
					id FROM currencies
				WHERE
					name = 'EUR'::text))
	GROUP BY
		year) eur ON timetable.year = eur.year
	LEFT JOIN (
		SELECT
			COALESCE(sum(liabilities.total), 0::numeric) AS gbp, date_part('year', liabilities.created_at) AS year
		FROM
			liabilities
		WHERE (liabilities.paid_at IS NULL)
		AND(liabilities.currency_id IN(
				SELECT
					id FROM currencies
				WHERE
					name = 'GBP'::text))
	GROUP BY
		year) gbp ON timetable.year = gbp.year
	LEFT JOIN (
		SELECT
			COALESCE(sum(liabilities.total), 0::numeric) AS sgd, date_part('year', liabilities.created_at) AS year
		FROM
			liabilities
		WHERE (liabilities.paid_at IS NULL)
		AND(liabilities.currency_id IN(
				SELECT
					id FROM currencies
				WHERE
					name = 'SGD'::text))
	GROUP BY
		year) sgd ON timetable.year = sgd.year
ORDER BY
	timetable.year DESC;


-- //////////////////////////////////////
-- VIEW INCOMES
-- //////////////////////////////////////
create or replace view vw_incomes AS
SELECT
	COALESCE(
		vnd.vnd, 0::numeric::double precision) AS vnd,
	COALESCE(
		usd.usd, 0::numeric::double precision) AS usd,
	COALESCE(
		eur.eur, 0::numeric::double precision) AS eur,
	COALESCE(
		gbp.gbp, 0::numeric::double precision) AS gbp,
	COALESCE(
		sgd.sgd, 0::numeric::double precision) AS sgd,
	timetable.month,
	timetable.year
FROM (
	SELECT
		date_part('month'::text, accounting_transactions.date) AS month,
		date_part('year'::text, accounting_transactions.date) AS year
	FROM
		accounting_transactions
	GROUP BY
		(date_part('month'::text, accounting_transactions.date)),
		(date_part('year'::text, accounting_transactions.date))) timetable
	LEFT JOIN (
		SELECT
			COALESCE(sum(t.amount), 0::numeric::double precision) AS vnd,
			date_part('year'::text, t.date) AS year,
			date_part('month'::text, t.date) AS month
		FROM
			accounting_transactions t
		WHERE
			t.category = 'In' ::text
			AND t.currency = 'VND' ::text
		GROUP BY
			(date_part('month'::text, t.date)),
			(date_part('year'::text, t.date))) vnd ON timetable.month = vnd.month
	AND timetable.year = vnd.year
	LEFT JOIN (
		SELECT
			COALESCE(sum(t.amount), 0::numeric::double precision) AS usd,
			date_part('year'::text, t.date) AS year,
			date_part('month'::text, t.date) AS month
		FROM
			accounting_transactions t
		WHERE
			t.category = 'In' ::text
			AND t.currency = 'USD' ::text
		GROUP BY
			(date_part('month'::text, t.date)),
			(date_part('year'::text, t.date))) usd ON timetable.month = usd.month
	AND timetable.year = usd.year
	LEFT JOIN (
		SELECT
			COALESCE(sum(t.amount), 0::numeric::double precision) AS eur,
			date_part('year'::text, t.date) AS year,
			date_part('month'::text, t.date) AS month
		FROM
			accounting_transactions t
		WHERE
			t.category = 'In' ::text
			AND t.currency = 'EUR' ::text
		GROUP BY
			(date_part('month'::text, t.date)),
			(date_part('year'::text, t.date))) eur ON timetable.month = eur.month
	AND timetable.year = eur.year
	LEFT JOIN (
		SELECT
			COALESCE(sum(t.amount), 0::numeric::double precision) AS gbp,
			date_part('year'::text, t.date) AS year,
			date_part('month'::text, t.date) AS month
		FROM
			accounting_transactions t
		WHERE
			t.category = 'In' ::text
			AND t.currency = 'GBP' ::text
		GROUP BY
			(date_part('month'::text, t.date)),
			(date_part('year'::text, t.date))) gbp ON timetable.month = gbp.month
	AND timetable.year = gbp.year
	LEFT JOIN (
		SELECT
			COALESCE(sum(t.amount), 0::numeric::double precision) AS sgd,
			date_part('year'::text, t.date) AS year,
			date_part('month'::text, t.date) AS month
		FROM
			accounting_transactions t
		WHERE
			t.category = 'In' ::text
			AND t.currency = 'SGD' ::text
		GROUP BY
			(date_part('month'::text, t.date)),
			(date_part('year'::text, t.date))) sgd ON timetable.month = sgd.month
	AND timetable.year = sgd.year
ORDER BY
	timetable.year DESC,
	timetable.month DESC;

create or replace view vw_expenses as
SELECT sum(t.conversion_amount) AS vnd,
	date_part('year'::text, t.date) AS year
	FROM accounting_transactions t
WHERE (t.category <> 'In'::text OR t.category IS NULL) AND t.category !~~ '%Payroll%'::text AND t.category !~~ '%Commission%'::text AND t.category !~~ '%Investment%'::text
GROUP BY (date_part('year'::text, t.date))
ORDER BY (date_part('year'::text, t.date)) DESC;

create or replace view vw_payrolls as
 SELECT COALESCE(total.total, 0::bigint) AS total,
    COALESCE(jan.jan, 0::bigint) AS jan,
    COALESCE(feb.feb, 0::bigint) AS feb,
    COALESCE(mar.mar, 0::bigint) AS mar,
    COALESCE(apr.apr, 0::bigint) AS apr,
    COALESCE(may.may, 0::bigint) AS may,
    COALESCE(jun.jun, 0::bigint) AS jun,
    COALESCE(jul.jul, 0::bigint) AS jul,
    COALESCE(aug.aug, 0::bigint) AS aug,
    COALESCE(sep.sep, 0::bigint) AS sep,
    COALESCE(oct.oct, 0::bigint) AS oct,
    COALESCE(nov.nov, 0::bigint) AS nov,
    COALESCE("dec"."dec", 0::bigint) AS "dec",
    timetable.year
   FROM ( SELECT DISTINCT date_part('year'::text, accounting_transactions.date) AS year
           FROM accounting_transactions) timetable
     LEFT JOIN ( SELECT COALESCE(sum(t.conversion_amount), 0::numeric)::bigint AS total,
            date_part('year'::text, t.date) AS year
           FROM accounting_transactions t
          WHERE t.category ~~ 'Payroll %'::text OR t.category ~~ '%Commission%'::text
          GROUP BY (date_part('year'::text, t.date))) total ON timetable.year = total.year
     LEFT JOIN ( SELECT COALESCE(sum(t.conversion_amount), 0::numeric)::bigint AS jan,
            date_part('year'::text, t.date) AS year
           FROM accounting_transactions t
          WHERE (t.category ~~ 'Payroll %'::text OR t.category ~~ '%Commission%'::text) AND date_part('month'::text, t.date) = 1::double precision
          GROUP BY (date_part('year'::text, t.date))) jan ON timetable.year = jan.year
     LEFT JOIN ( SELECT COALESCE(sum(t.conversion_amount), 0::numeric)::bigint AS feb,
            date_part('year'::text, t.date) AS year
           FROM accounting_transactions t
          WHERE (t.category ~~ 'Payroll %'::text OR t.category ~~ '%Commission%'::text) AND date_part('month'::text, t.date) = 2::double precision
          GROUP BY (date_part('year'::text, t.date))) feb ON timetable.year = feb.year
     LEFT JOIN ( SELECT COALESCE(sum(t.conversion_amount), 0::numeric)::bigint AS mar,
            date_part('year'::text, t.date) AS year
           FROM accounting_transactions t
          WHERE (t.category ~~ 'Payroll %'::text OR t.category ~~ '%Commission%'::text) AND date_part('month'::text, t.date) = 3::double precision
          GROUP BY (date_part('year'::text, t.date))) mar ON timetable.year = mar.year
     LEFT JOIN ( SELECT COALESCE(sum(t.conversion_amount), 0::numeric)::bigint AS apr,
            date_part('year'::text, t.date) AS year
           FROM accounting_transactions t
          WHERE (t.category ~~ 'Payroll %'::text OR t.category ~~ '%Commission%'::text) AND date_part('month'::text, t.date) = 4::double precision
          GROUP BY (date_part('year'::text, t.date))) apr ON timetable.year = apr.year
     LEFT JOIN ( SELECT COALESCE(sum(t.conversion_amount), 0::numeric)::bigint AS may,
            date_part('year'::text, t.date) AS year
           FROM accounting_transactions t
          WHERE (t.category ~~ 'Payroll %'::text OR t.category ~~ '%Commission%'::text) AND date_part('month'::text, t.date) = 5::double precision
          GROUP BY (date_part('year'::text, t.date))) may ON timetable.year = may.year
     LEFT JOIN ( SELECT COALESCE(sum(t.conversion_amount), 0::numeric)::bigint AS jun,
            date_part('year'::text, t.date) AS year
           FROM accounting_transactions t
          WHERE (t.category ~~ 'Payroll %'::text OR t.category ~~ '%Commission%'::text) AND date_part('month'::text, t.date) = 6::double precision
          GROUP BY (date_part('year'::text, t.date))) jun ON timetable.year = jun.year
     LEFT JOIN ( SELECT COALESCE(sum(t.conversion_amount), 0::numeric)::bigint AS jul,
            date_part('year'::text, t.date) AS year
           FROM accounting_transactions t
          WHERE (t.category ~~ 'Payroll %'::text OR t.category ~~ '%Commission%'::text) AND date_part('month'::text, t.date) = 7::double precision
          GROUP BY (date_part('year'::text, t.date))) jul ON timetable.year = jul.year
     LEFT JOIN ( SELECT COALESCE(sum(t.conversion_amount), 0::numeric)::bigint AS aug,
            date_part('year'::text, t.date) AS year
           FROM accounting_transactions t
          WHERE (t.category ~~ 'Payroll %'::text OR t.category ~~ '%Commission%'::text) AND date_part('month'::text, t.date) = 8::double precision
          GROUP BY (date_part('year'::text, t.date))) aug ON timetable.year = aug.year
     LEFT JOIN ( SELECT COALESCE(sum(t.conversion_amount), 0::numeric)::bigint AS sep,
            date_part('year'::text, t.date) AS year
           FROM accounting_transactions t
          WHERE (t.category ~~ 'Payroll %'::text OR t.category ~~ '%Commission%'::text) AND date_part('month'::text, t.date) = 9::double precision
          GROUP BY (date_part('year'::text, t.date))) sep ON timetable.year = sep.year
     LEFT JOIN ( SELECT COALESCE(sum(t.conversion_amount), 0::numeric)::bigint AS oct,
            date_part('year'::text, t.date) AS year
           FROM accounting_transactions t
          WHERE (t.category ~~ 'Payroll %'::text OR t.category ~~ '%Commission%'::text) AND date_part('month'::text, t.date) = 10::double precision
          GROUP BY (date_part('year'::text, t.date))) oct ON timetable.year = oct.year
     LEFT JOIN ( SELECT COALESCE(sum(t.conversion_amount), 0::numeric)::bigint AS nov,
            date_part('year'::text, t.date) AS year
           FROM accounting_transactions t
          WHERE (t.category ~~ 'Payroll %'::text OR t.category ~~ '%Commission%'::text) AND date_part('month'::text, t.date) = 11::double precision
          GROUP BY (date_part('year'::text, t.date))) nov ON timetable.year = nov.year
     LEFT JOIN ( SELECT COALESCE(sum(t.conversion_amount), 0::numeric)::bigint AS "dec",
            date_part('year'::text, t.date) AS year
           FROM accounting_transactions t
          WHERE (t.category ~~ 'Payroll %'::text OR t.category ~~ '%Commission%'::text) AND date_part('month'::text, t.date) = 12::double precision
          GROUP BY (date_part('year'::text, t.date))) "dec" ON timetable.year = "dec".year
  ORDER BY timetable.year DESC;

create or replace view vw_investments as
 SELECT
	sum(t.conversion_amount) AS vnd,
	date_part('year'::text, t.date) AS year
FROM
	accounting_transactions t
WHERE (t.category = 'Investment'::text)
GROUP BY
	(date_part('year'::text, t.date))
ORDER BY
	(date_part('year'::text, t.date))
	DESC;

-- +migrate Down
drop view if exists vw_expenses;
drop view if exists vw_incomes;
drop view if exists vw_account_receivables;
drop view if exists vw_payrolls;
drop view if exists vw_liabilities;
drop view if exists vw_investments;
