
-- +migrate Up
CREATE TABLE physical_checkin_transactions (
    id SERIAL PRIMARY KEY,
    employee_id UUID REFERENCES employees(id) NOT NULL,
    date DATE NOT NULL,
    icy_amount FLOAT8 DEFAULT 0.0,
    mochi_tx_id INTEGER
);
-- add unique constraint
CREATE UNIQUE INDEX physical_checkin_transactions_employee_id_date_idx ON physical_checkin_transactions(employee_id, date);


-- +migrate Down
DROP TABLE physical_checkin_transactions;


