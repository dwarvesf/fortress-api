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
            unnest(to_tsvector(LOWER(c."full_name" || ' ' || c."team_email" || ' ' || c."discord_id" || ' ' || c."notion_id" || ' ' || c."github_id" ))),
            generate_series(1, length(lexeme)) len));

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_insert_keyword_vector ()
	RETURNS TRIGGER
	AS $BODY$
BEGIN
  UPDATE
        "employees" c
      SET
        "keyword_vector" = array_to_tsvector ((
            SELECT
              array_agg(DISTINCT substring(lexeme FOR len))
            FROM
              unnest(to_tsvector(LOWER(c."full_name" || ' ' || c."team_email" || ' ' || c."discord_id" || ' ' || c."notion_id" || ' ' || c."github_id" ))),
              generate_series(1, length(lexeme)) len)) WHERE c.id = NEW.id;
      RETURN new;
END;
$BODY$
LANGUAGE plpgsql;
-- +migrate StatementEnd

CREATE TRIGGER trig_insert_keyword_vector
     AFTER INSERT ON employees
     FOR EACH ROW
     EXECUTE PROCEDURE fn_insert_keyword_vector();


-- +migrate Down

DROP TRIGGER IF EXISTS trig_insert_keyword_vector ON employees;
DROP FUNCTION fn_insert_keyword_vector;

ALTER TABLE "employees" DROP COLUMN "keyword_vector";
DROP TEXT SEARCH CONFIGURATION  IF EXISTS public.english_nostop;
DROP TEXT SEARCH DICTIONARY  IF EXISTS  english_stem_nostop;
