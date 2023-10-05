-- +migrate Up
CREATE TABLE IF NOT EXISTS employee_mma_scores (
    id              UUID PRIMARY KEY DEFAULT (UUID()),
    deleted_at      TIMESTAMP(6),
    created_at      TIMESTAMP(6) DEFAULT (now()),
    updated_at      TIMESTAMP(6) DEFAULT (now()),

    employee_id     UUID DEFAULT NULL,
    mastery_score   DECIMAL,
    autonomy_score  DECIMAL,
    meaning_score   DECIMAL,
    rated_at        TIMESTAMP(6) DEFAULT (now())
);

ALTER TABLE employee_mma_scores
    ADD CONSTRAINT employee_mma_scores_employee_id_fkey FOREIGN KEY (employee_id) REFERENCES employees (id);

-- +migrate Down
DROP TABLE IF EXISTS employee_mma_scores;
