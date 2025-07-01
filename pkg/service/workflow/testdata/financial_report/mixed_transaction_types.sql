-- Test data for mixed transaction types
-- Tests that only expense types (SE, OP, OV, CA) are included in expense calculation

INSERT INTO accounting_transactions (id, type, name, date, amount, conversion_amount, conversion_rate, created_at, updated_at)
VALUES 
-- Expense transactions that SHOULD be included (120,000 VND total)
('txn-001', 'SE', 'Salary Expense', '2025-06-15', 2.0, 30000, 25000, NOW(), NOW()),
('txn-002', 'OP', 'Office Expense', '2025-06-15', 1.6, 40000, 25000, NOW(), NOW()),
('txn-003', 'OV', 'Overhead Expense', '2025-06-20', 2.0, 50000, 25000, NOW(), NOW()),
-- Income transactions that should NOT be included
('txn-004', 'In', 'Income Transaction', '2025-06-15', 4.0, 100000, 25000, NOW(), NOW()),
('txn-005', 'In', 'Another Income', '2025-06-20', 2.0, 50000, 25000, NOW(), NOW()),
-- Other transaction types that should NOT be included  
('txn-006', 'AS', 'Asset Transaction', '2025-06-15', 3.0, 75000, 25000, NOW(), NOW()),
('txn-007', 'LI', 'Liability Transaction', '2025-06-20', 1.0, 25000, 25000, NOW(), NOW());

-- Expected total expenses: 30,000 + 40,000 + 50,000 = 120,000 VND