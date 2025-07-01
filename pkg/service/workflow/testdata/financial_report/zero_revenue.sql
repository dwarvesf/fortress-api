-- Test data for zero revenue scenario - expenses but no revenue
-- This creates test data with employees and expenses but no paid invoices

-- Insert test employee
INSERT INTO employees (id, full_name, username, team_email, personal_email, avatar, working_status, left_date, joined_date, created_at, updated_at)
VALUES 
('2655832e-f009-4b73-a535-64c3a22e558f', 'John Doe', 'john.doe', 'john.doe@d.foundation', 'john@gmail.com', '', 'full-time', NULL, '2023-01-01', NOW(), NOW());

-- Insert test project (inactive to make employee non-billable)
INSERT INTO projects (id, name, status, start_date, created_at, updated_at)
VALUES 
('8dc3be2e-19a4-4942-8a79-56db391a0b15', 'Test Project', 'inactive', '2023-01-01', NOW(), NOW());

-- Insert expense transactions for June 2025 (50,000 VND total)
INSERT INTO accounting_transactions (id, type, name, date, amount, conversion_amount, conversion_rate, created_at, updated_at)
VALUES 
('txn-001', 'SE', 'Salary Expense', '2025-06-15', 2.0, 50000, 25000, NOW(), NOW());

-- Insert YTD income transaction from previous months (100,000 VND)
INSERT INTO accounting_transactions (id, type, name, date, amount, conversion_amount, conversion_rate, created_at, updated_at)
VALUES 
('txn-002', 'In', 'YTD Income', '2025-01-15', 4.0, 100000, 25000, NOW(), NOW());

-- No paid invoices in June = zero revenue