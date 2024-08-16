-- +migrate Up
CREATE TABLE IF NOT EXISTS invoice_emails (
    id SERIAL PRIMARY KEY,
    sender VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    received_at TIMESTAMP NOT NULL,
    invoice_number VARCHAR(100),
    invoice_amount DECIMAL(10, 2),
    invoice_date DATE,
    content TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE IF EXISTS invoice_emails;