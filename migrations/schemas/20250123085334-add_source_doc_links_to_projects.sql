-- +migrate Up
ALTER TABLE projects 
  ADD COLUMN source_link VARCHAR(255),
  ADD COLUMN doc_link VARCHAR(255);

-- +migrate Down
ALTER TABLE projects 
  DROP COLUMN source_link,
  DROP COLUMN doc_link;
