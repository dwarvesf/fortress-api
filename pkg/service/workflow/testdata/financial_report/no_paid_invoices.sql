-- Test data for no paid invoices scenario
-- Creates unpaid invoices but no paid ones for December 2025

INSERT INTO invoices (id, number, email, total, subtotal, tax, discount, status, created_at, updated_at, paid_at, conversion_amount)
VALUES 
-- Unpaid invoices for December 2025
('invoice-001', 'INV-2025-100', 'client1@example.com', 4.0, 4.0, 0, 0, 'sent', '2025-12-15 10:00:00', '2025-12-15 10:00:00', NULL, 100000.0),
('invoice-002', 'INV-2025-101', 'client2@example.com', 6.0, 6.0, 0, 0, 'draft', '2025-12-20 10:00:00', '2025-12-20 10:00:00', NULL, 150000.0),
-- Paid invoice from different month
('invoice-003', 'INV-2025-102', 'client3@example.com', 8.0, 8.0, 0, 0, 'paid', '2025-11-15 10:00:00', '2025-11-15 10:00:00', '2025-11-15 10:00:00', 200000.0);