-- Test data for currency conversion accuracy testing
-- This creates test data with known VND amounts for precise conversion testing

-- Insert 2 test employees
INSERT INTO employees (id, full_name, username, team_email, personal_email, avatar, working_status, left_date, joined_date, created_at, updated_at)
VALUES 
('employee-001', 'John Doe', 'john.doe', 'john.doe@d.foundation', 'john@gmail.com', '', 'full-time', NULL, '2023-01-01', NOW(), NOW()),
('employee-002', 'Jane Smith', 'jane.smith', 'jane.smith@d.foundation', 'jane@gmail.com', '', 'full-time', NULL, '2023-01-01', NOW(), NOW());

-- Insert test project
INSERT INTO projects (id, name, status, start_date, created_at, updated_at)
VALUES 
('project-001', 'Test Project', 'active', '2023-01-01', NOW(), NOW());

-- Insert project members (makes employees billable)
INSERT INTO project_members (id, project_id, employee_id, positions, deployment_type, rate, discount, status, start_date, end_date, created_at, updated_at)
VALUES 
('member-001', 'project-001', 'employee-001', '[]', 'full-time', 0, 0, 'active', '2023-01-01', NULL, NOW(), NOW()),
('member-002', 'project-001', 'employee-002', '[]', 'full-time', 0, 0, 'active', '2023-01-01', NULL, NOW(), NOW());

-- Insert paid invoices for June 2025 (500,000 VND total)
INSERT INTO invoices (id, number, email, total, subtotal, tax, discount, status, created_at, updated_at, paid_at, conversion_amount)
VALUES 
('invoice-001', 'INV-2025-001', 'client@example.com', 12.0, 12.0, 0, 0, 'paid', '2025-06-15 10:00:00', '2025-06-15 10:00:00', '2025-06-15 10:00:00', 300000.0),
('invoice-002', 'INV-2025-002', 'client2@example.com', 8.0, 8.0, 0, 0, 'paid', '2025-06-20 10:00:00', '2025-06-20 10:00:00', '2025-06-20 10:00:00', 200000.0);

-- Insert expense transactions for June 2025 (250,000 VND total)
INSERT INTO accounting_transactions (id, type, name, date, amount, conversion_amount, conversion_rate, created_at, updated_at)
VALUES 
('txn-001', 'SE', 'Salary Expense', '2025-06-15', 6.0, 150000, 25000, NOW(), NOW()),
('txn-002', 'OP', 'Office Expense', '2025-06-15', 4.0, 100000, 25000, NOW(), NOW());

-- Insert YTD income transaction (400,000 VND)
INSERT INTO accounting_transactions (id, type, name, date, amount, conversion_amount, conversion_rate, created_at, updated_at)
VALUES 
('txn-003', 'In', 'YTD Income', '2025-01-15', 16.0, 400000, 25000, NOW(), NOW());

-- Insert unpaid invoice for receivables (100,000 VND)
INSERT INTO invoices (id, number, email, total, subtotal, tax, discount, status, created_at, updated_at, paid_at, conversion_amount)
VALUES 
('invoice-003', 'INV-2025-003', 'client3@example.com', 4.0, 4.0, 0, 0, 'sent', '2025-06-25 10:00:00', '2025-06-25 10:00:00', NULL, 100000.0);