-- +migrate Up
ALTER TABLE "employees" ADD COLUMN "keyword_vector" tsvector;

CREATE TEXT SEARCH DICTIONARY english_stem_nostop (
    Template = snowball
    , Language = english
);

CREATE TEXT SEARCH CONFIGURATION public.english_nostop ( COPY = pg_catalog.english );

ALTER TEXT SEARCH CONFIGURATION public.english_nostop
   ALTER MAPPING FOR asciiword, asciihword, hword_asciipart, hword, hword_part, word WITH english_stem_nostop;

CREATE INDEX idx_employees_keyword_vector ON "employees" USING gin(keyword_vector);

UPDATE
      "employees" c
    SET
      "keyword_vector" = array_to_tsvector ((
          SELECT
            array_agg(DISTINCT substring(lexeme FOR len))
          FROM
            unnest(to_tsvector(LOWER(COALESCE(c. "full_name", '') || ' ' || COALESCE(c. "team_email", '') || ' ' || COALESCE(c. "discord_id", '') || ' ' || COALESCE(c. "notion_id", '') || ' ' || COALESCE(c. "github_id", '') || ' ' || COALESCE(c. "notion_name", '') || ' ' || COALESCE(c. "discord_name", '')))),
            generate_series(1, length(lexeme)) len));

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
				unnest(to_tsvector(LOWER(COALESCE(NEW. "full_name", '') || ' ' || COALESCE(NEW. "team_email", '') || ' ' || COALESCE(NEW. "discord_id", '') || ' ' || COALESCE(NEW. "notion_id", '') || ' ' || COALESCE(NEW. "github_id", '') || ' ' || COALESCE(NEW. "notion_name", '') || ' ' || COALESCE(NEW. "discord_name", '')))),
				generate_series(1, length(lexeme)) len));
	RETURN NEW;
END;
$function$
-- +migrate StatementEnd

CREATE TRIGGER trig_insert_keyword_vector
     BEFORE INSERT OR UPDATE ON employees
     FOR EACH ROW
     EXECUTE PROCEDURE fn_insert_keyword_vector();


-- +migrate Down

DROP TRIGGER IF EXISTS trig_insert_keyword_vector ON employees;
DROP FUNCTION fn_insert_keyword_vector;

ALTER TABLE "employees" DROP COLUMN "keyword_vector";
DROP TEXT SEARCH CONFIGURATION  IF EXISTS public.english_nostop;
DROP TEXT SEARCH DICTIONARY  IF EXISTS  english_stem_nostop;
