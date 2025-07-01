-- Test data for zero division protection - no employees scenario
-- This creates test data with no active employees to test division by zero protection

-- Insert some invoices and transactions but no employees
INSERT INTO invoices (id, number, email, total, sub_total, tax, discount, status, created_at, updated_at, paid_at, conversion_amount)
VALUES 
('550e8400-e29b-41d4-a716-446655440010', 'INV-2025-001', 'client@example.com', 4.0, 4.0, 0, 0, 'paid', '2025-06-15 10:00:00', '2025-06-15 10:00:00', '2025-06-15 10:00:00', 100000.0);

INSERT INTO accounting_transactions (id, type, name, date, amount, conversion_amount, conversion_rate, created_at, updated_at)
VALUES 
('550e8400-e29b-41d4-a716-446655440011', 'SE', 'Salary Expense', '2025-06-15', 2.0, 50000, 25000, NOW(), NOW());