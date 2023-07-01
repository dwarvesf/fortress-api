-- +migrate Up
DROP VIEW vw_employees_recently_joined;

ALTER TABLE employees DROP COLUMN gitlab_id;
ALTER TABLE employees DROP COLUMN github_id;
ALTER TABLE employees DROP COLUMN discord_id;
ALTER TABLE employees DROP COLUMN notion_id;
ALTER TABLE employees DROP COLUMN discord_name;
ALTER TABLE employees DROP COLUMN notion_name;
ALTER TABLE employees DROP COLUMN notion_email;
ALTER TABLE employees DROP COLUMN linkedin_name;

CREATE OR REPLACE VIEW vw_employees_recently_joined AS
SELECT *
FROM employees
WHERE joined_date BETWEEN CURRENT_DATE - INTERVAL '7 days' AND CURRENT_DATE;

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION public.fn_insert_keyword_vector ()
	RETURNS TRIGGER
	LANGUAGE plpgsql
	AS $function$
BEGIN
	NEW. "keyword_vector" = array_to_tsvector ((
			SELECT
				array_agg(DISTINCT substring(lexeme FOR len))
			FROM
				unnest(
				    to_tsvector(
				        fn_remove_vietnamese_accents(
				            LOWER(COALESCE(NEW. "full_name", '') || ' ' || COALESCE(NEW. "team_email", ''))
				        )
				    )
				),
				generate_series(1, length(lexeme)) len));
RETURN NEW;
END;
$function$;
-- +migrate StatementEnd
UPDATE
    "employees" c
SET "keyword_vector" = array_to_tsvector ((
    SELECT
        array_agg(DISTINCT substring(lexeme FOR len))
    FROM
        unnest(
                to_tsvector(
                        fn_remove_vietnamese_accents(
                                LOWER(COALESCE(c. "full_name", '') || ' ' || COALESCE(c. "team_email", ''))
                            )
                    )
            ),
        generate_series(1, length(lexeme)) len))
WHERE TRUE;

-- +migrate Down
ALTER TABLE employees ADD COLUMN gitlab_id TEXT;
ALTER TABLE employees ADD COLUMN github_id TEXT;
ALTER TABLE employees ADD COLUMN discord_id TEXT;
ALTER TABLE employees ADD COLUMN notion_id TEXT;
ALTER TABLE employees ADD COLUMN discord_name TEXT;
ALTER TABLE employees ADD COLUMN notion_name TEXT;
ALTER TABLE employees ADD COLUMN notion_email TEXT;
ALTER TABLE employees ADD COLUMN linkedin_name TEXT;

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION public.fn_insert_keyword_vector ()
	RETURNS TRIGGER
	LANGUAGE plpgsql
	AS $function$
BEGIN
    NEW. "keyword_vector" = array_to_tsvector ((
        SELECT
            array_agg(DISTINCT substring(lexeme FOR len))
        FROM
            unnest(
                to_tsvector(
                    fn_remove_vietnamese_accents(
                        LOWER(COALESCE(NEW. "full_name", '') || ' ' || COALESCE(NEW. "team_email", '') || ' ' || COALESCE(NEW. "discord_id", '') || ' ' || COALESCE(NEW. "notion_id", '') || ' ' || COALESCE(NEW. "github_id", '') || ' ' || COALESCE(NEW. "notion_name", '') || ' ' || COALESCE(NEW. "discord_name", ''))
                    )
                )
            ),
            generate_series(1, length(lexeme)) len));
RETURN NEW;
END;
$function$;
-- +migrate StatementEnd
UPDATE
    "employees" c
SET
    "keyword_vector" = array_to_tsvector ((
        SELECT
            array_agg(DISTINCT substring(lexeme FOR len))
        FROM
            unnest(
                    to_tsvector(
                            fn_remove_vietnamese_accents(
                                    LOWER(COALESCE(c. "full_name", '') || ' ' || COALESCE(c. "team_email", '') || ' ' || COALESCE(c. "discord_id", '') || ' ' || COALESCE(c. "notion_id", '') || ' ' || COALESCE(c. "github_id", '') || ' ' || COALESCE(c. "notion_name", '') || ' ' || COALESCE(c. "discord_name", ''))
                                )
                        )
                ),
            generate_series(1, length(lexeme)) len))
WHERE TRUE;
