-- Test data for successful financial report generation
-- This creates minimal test data with 1 employee, 1 paid invoice, and expenses

-- Insert test employee with all required fields
INSERT INTO employees (
    id, full_name, display_name, username, team_email, personal_email, 
    avatar, gender, working_status, joined_date, created_at, updated_at
)
VALUES (
    '2655832e-f009-4b73-a535-64c3a22e558f', 
    'John Doe', 
    'John Doe', 
    'john.doe', 
    'john.doe@d.foundation', 
    'john@gmail.com', 
    '', 
    'male', 
    'full-time', 
    '2023-01-01', 
    NOW(), 
    NOW()
);

-- Insert test project
INSERT INTO projects (id, name, status, start_date, created_at, updated_at)
VALUES (
    '8dc3be2e-19a4-4942-8a79-56db391a0b15', 
    'Test Project', 
    'active', 
    '2023-01-01', 
    NOW(), 
    NOW()
);

-- Insert bank account (required for invoices)
INSERT INTO bank_accounts (id, account_number, bank_name, currency_id, owner_name, created_at, updated_at)
VALUES (
    '550e8400-e29b-41d4-a716-446655440008',
    '123456789',
    'Test Bank',
    '550e8400-e29b-41d4-a716-446655440009',
    'Test Company',
    NOW(),
    NOW()
);

-- Insert currency
INSERT INTO currencies (id, name, symbol, locale, created_at, updated_at)
VALUES (
    '550e8400-e29b-41d4-a716-446655440009',
    'Vietnamese Dong',
    'VND',
    'vi-VN',
    NOW(),
    NOW()
);

-- Insert project slot (required for project members)
INSERT INTO project_slots (id, project_id, deployment_type, status, created_at, updated_at)
VALUES (
    '550e8400-e29b-41d4-a716-446655440012',
    '8dc3be2e-19a4-4942-8a79-56db391a0b15',
    'official',
    'active',
    NOW(),
    NOW()
);

-- Insert seniority (required for project members)
INSERT INTO seniorities (id, name, created_at, updated_at)
VALUES (
    '550e8400-e29b-41d4-a716-446655440013',
    'Senior',
    NOW(),
    NOW()
);

-- Insert project member (makes employee billable)
INSERT INTO project_members (
    id, project_id, project_slot_id, employee_id, seniority_id, 
    joined_date, rate, discount, status, deployment_type, created_at, updated_at
)
VALUES (
    '550e8400-e29b-41d4-a716-446655440001', 
    '8dc3be2e-19a4-4942-8a79-56db391a0b15', 
    '550e8400-e29b-41d4-a716-446655440012',
    '2655832e-f009-4b73-a535-64c3a22e558f', 
    '550e8400-e29b-41d4-a716-446655440013',
    '2023-01-01', 
    0, 
    0, 
    'active', 
    'official', 
    NOW(), 
    NOW()
);

-- Insert paid invoice for June 2025 (100,000 VND)
INSERT INTO invoices (
    id, number, email, total, sub_total, tax, discount, status, 
    project_id, bank_id, sent_by, created_at, updated_at, paid_at, conversion_amount
)
VALUES (
    '550e8400-e29b-41d4-a716-446655440002', 
    'INV-2025-001', 
    'client@example.com', 
    4.0, 
    4, 
    0, 
    0, 
    'paid', 
    '8dc3be2e-19a4-4942-8a79-56db391a0b15',
    '550e8400-e29b-41d4-a716-446655440008',
    '2655832e-f009-4b73-a535-64c3a22e558f',
    '2025-06-15 10:00:00', 
    '2025-06-15 10:00:00', 
    '2025-06-15 10:00:00', 
    100000
);

-- Insert expense transactions for June 2025 (50,000 VND total)
INSERT INTO accounting_transactions (id, type, name, date, amount, conversion_amount, conversion_rate, created_at, updated_at)
VALUES 
('550e8400-e29b-41d4-a716-446655440003', 'SE', 'Salary Expense', '2025-06-15', 2.0, 25000, 25000, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440004', 'OP', 'Office Expense', '2025-06-15', 1.0, 25000, 25000, NOW(), NOW());

-- Insert YTD income transaction (200,000 VND)
INSERT INTO accounting_transactions (id, type, name, date, amount, conversion_amount, conversion_rate, created_at, updated_at)
VALUES 
('550e8400-e29b-41d4-a716-446655440005', 'In', 'YTD Income', '2025-01-15', 8.0, 100000, 25000, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440006', 'In', 'YTD Income 2', '2025-03-15', 4.0, 100000, 25000, NOW(), NOW());

-- Insert unpaid invoice for receivables (50,000 VND)
INSERT INTO invoices (
    id, number, email, total, sub_total, tax, discount, status, 
    project_id, bank_id, sent_by, created_at, updated_at, paid_at, conversion_amount
)
VALUES (
    '550e8400-e29b-41d4-a716-446655440007', 
    'INV-2025-002', 
    'client2@example.com', 
    2.0, 
    2, 
    0, 
    0, 
    'sent', 
    '8dc3be2e-19a4-4942-8a79-56db391a0b15',
    '550e8400-e29b-41d4-a716-446655440008',
    '2655832e-f009-4b73-a535-64c3a22e558f',
    '2025-06-20 10:00:00', 
    '2025-06-20 10:00:00', 
    NULL, 
    50000
);