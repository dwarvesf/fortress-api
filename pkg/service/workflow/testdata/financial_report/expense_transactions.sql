-- Test data for expense transactions calculation
-- Creates various expense transaction types to test expense aggregation

INSERT INTO accounting_transactions (id, type, name, date, amount, conversion_amount, conversion_rate, created_at, updated_at)
VALUES 
-- Expense transactions that should be included (SE, OP, OV, CA)
('txn-001', 'SE', 'Salary Expense', '2025-06-15', 2.0, 50000, 25000, NOW(), NOW()),
('txn-002', 'OP', 'Office Expense', '2025-06-15', 1.6, 40000, 25000, NOW(), NOW()),
('txn-003', 'OV', 'Overhead Expense', '2025-06-20', 2.4, 60000, 25000, NOW(), NOW()),
('txn-004', 'CA', 'Capital Expense', '2025-06-25', 1.2, 30000, 25000, NOW(), NOW()),
-- Income transaction should not be included
('txn-005', 'In', 'Income Transaction', '2025-06-15', 4.0, 100000, 25000, NOW(), NOW()),
-- Expense from different month should not be included
('txn-006', 'SE', 'Previous Month Salary', '2025-05-15', 2.0, 50000, 25000, NOW(), NOW());