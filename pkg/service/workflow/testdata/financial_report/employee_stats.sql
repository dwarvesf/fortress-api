-- Test data for employee statistics calculation
-- Creates employees with different statuses and project assignments

-- Insert 5 full-time employees
INSERT INTO employees (id, full_name, username, team_email, personal_email, avatar, working_status, left_date, joined_date, created_at, updated_at)
VALUES 
('emp-001', 'John Doe', 'john.doe', 'john.doe@d.foundation', 'john@gmail.com', '', 'full-time', NULL, '2023-01-01', NOW(), NOW()),
('emp-002', 'Jane Smith', 'jane.smith', 'jane.smith@d.foundation', 'jane@gmail.com', '', 'full-time', NULL, '2023-01-01', NOW(), NOW()),
('emp-003', 'Bob Johnson', 'bob.johnson', 'bob.johnson@d.foundation', 'bob@gmail.com', '', 'full-time', NULL, '2023-01-01', NOW(), NOW()),
('emp-004', 'Alice Brown', 'alice.brown', 'alice.brown@d.foundation', 'alice@gmail.com', '', 'full-time', NULL, '2023-01-01', NOW(), NOW()),
('emp-005', 'Charlie Wilson', 'charlie.wilson', 'charlie.wilson@d.foundation', 'charlie@gmail.com', '', 'full-time', NULL, '2023-01-01', NOW(), NOW()),
-- Part-time employee should not be counted
('emp-006', 'Part Timer', 'part.timer', 'part.timer@d.foundation', 'part@gmail.com', '', 'part-time', NULL, '2023-01-01', NOW(), NOW()),
-- Contractor should not be counted
('emp-007', 'Contractor', 'contractor', 'contractor@d.foundation', 'contractor@gmail.com', '', 'contractor', NULL, '2023-01-01', NOW(), NOW());

-- Insert active projects
INSERT INTO projects (id, name, status, start_date, created_at, updated_at)
VALUES 
('project-001', 'Active Project 1', 'active', '2023-01-01', NOW(), NOW()),
('project-002', 'Active Project 2', 'active', '2023-01-01', NOW(), NOW()),
('project-003', 'Inactive Project', 'inactive', '2023-01-01', NOW(), NOW());

-- Insert project members (3 employees on active projects = 3 billable)
INSERT INTO project_members (id, project_id, employee_id, positions, deployment_type, rate, discount, status, start_date, end_date, created_at, updated_at)
VALUES 
('member-001', 'project-001', 'emp-001', '[]', 'full-time', 0, 0, 'active', '2023-01-01', NULL, NOW(), NOW()),
('member-002', 'project-001', 'emp-002', '[]', 'full-time', 0, 0, 'active', '2023-01-01', NULL, NOW(), NOW()),
('member-003', 'project-002', 'emp-003', '[]', 'full-time', 0, 0, 'active', '2023-01-01', NULL, NOW(), NOW()),
-- Employee on inactive project should not be billable
('member-004', 'project-003', 'emp-004', '[]', 'full-time', 0, 0, 'active', '2023-01-01', NULL, NOW(), NOW());

-- emp-005 has no project assignment = not billable
-- Part-time and contractor employees should not affect counts