-- Test data for no expenses scenario  
-- Creates income transactions but no expense transactions for December 2025

INSERT INTO accounting_transactions (id, type, name, date, amount, conversion_amount, conversion_rate, created_at, updated_at)
VALUES 
-- Income transactions for December 2025 (should not be counted as expenses)
('txn-001', 'In', 'Income Transaction', '2025-12-15', 4.0, 100000, 25000, NOW(), NOW()),
('txn-002', 'In', 'Another Income', '2025-12-20', 6.0, 150000, 25000, NOW(), NOW()),
-- Expense transaction from different month
('txn-003', 'SE', 'Previous Month Expense', '2025-11-15', 2.0, 50000, 25000, NOW(), NOW());