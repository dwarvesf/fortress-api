-- Test data for mixed employee status scenario
-- Tests that only full-time employees are counted

-- Insert employees with different working statuses  
INSERT INTO employees (id, full_name, username, team_email, personal_email, avatar, working_status, left_date, joined_date, created_at, updated_at)
VALUES 
-- Full-time employees (should be counted: 3 total)
('emp-001', 'John Doe', 'john.doe', 'john.doe@d.foundation', 'john@gmail.com', '', 'full-time', NULL, '2023-01-01', NOW(), NOW()),
('emp-002', 'Jane Smith', 'jane.smith', 'jane.smith@d.foundation', 'jane@gmail.com', '', 'full-time', NULL, '2023-01-01', NOW(), NOW()),
('emp-003', 'Bob Johnson', 'bob.johnson', 'bob.johnson@d.foundation', 'bob@gmail.com', '', 'full-time', NULL, '2023-01-01', NOW(), NOW()),
-- Part-time employees (should NOT be counted)
('emp-004', 'Part Timer 1', 'part1', 'part1@d.foundation', 'part1@gmail.com', '', 'part-time', NULL, '2023-01-01', NOW(), NOW()),
('emp-005', 'Part Timer 2', 'part2', 'part2@d.foundation', 'part2@gmail.com', '', 'part-time', NULL, '2023-01-01', NOW(), NOW()),
-- Contractors (should NOT be counted)  
('emp-006', 'Contractor 1', 'contractor1', 'contractor1@d.foundation', 'contractor1@gmail.com', '', 'contractor', NULL, '2023-01-01', NOW(), NOW()),
-- Terminated employees (should NOT be counted due to deleted_at)
('emp-007', 'Former Employee', 'former', 'former@d.foundation', 'former@gmail.com', '', 'full-time', '2023-12-31', '2023-01-01', NOW(), NOW());

-- Soft delete the terminated employee
UPDATE employees SET deleted_at = NOW() WHERE id = 'emp-007';

-- Insert active projects
INSERT INTO projects (id, name, status, start_date, created_at, updated_at)
VALUES 
('project-001', 'Active Project 1', 'active', '2023-01-01', NOW(), NOW()),
('project-002', 'Active Project 2', 'active', '2023-01-01', NOW(), NOW());

-- Insert project members (only 2 full-time employees on active projects = 2 billable)
INSERT INTO project_members (id, project_id, employee_id, positions, deployment_type, rate, discount, status, start_date, end_date, created_at, updated_at)
VALUES 
('member-001', 'project-001', 'emp-001', '[]', 'full-time', 0, 0, 'active', '2023-01-01', NULL, NOW(), NOW()),
('member-002', 'project-002', 'emp-002', '[]', 'full-time', 0, 0, 'active', '2023-01-01', NULL, NOW(), NOW()),
-- Part-time employee on project (should not count as billable since not full-time)
('member-003', 'project-001', 'emp-004', '[]', 'part-time', 0, 0, 'active', '2023-01-01', NULL, NOW(), NOW());

-- emp-003 is full-time but not on any project = not billable
-- Expected: 3 total full-time employees, 2 billable full-time employees