-- Test data for paid invoices calculation
-- Creates multiple paid invoices to test revenue calculation

INSERT INTO invoices (id, number, email, total, subtotal, tax, discount, status, created_at, updated_at, paid_at, conversion_amount)
VALUES 
('invoice-001', 'INV-2025-001', 'client1@example.com', 4.0, 4.0, 0, 0, 'paid', '2025-06-15 10:00:00', '2025-06-15 10:00:00', '2025-06-15 10:00:00', 100000.0),
('invoice-002', 'INV-2025-002', 'client2@example.com', 6.0, 6.0, 0, 0, 'paid', '2025-06-20 10:00:00', '2025-06-20 10:00:00', '2025-06-20 10:00:00', 150000.0),
-- Unpaid invoice should not be included
('invoice-003', 'INV-2025-003', 'client3@example.com', 2.0, 2.0, 0, 0, 'sent', '2025-06-25 10:00:00', '2025-06-25 10:00:00', NULL, 50000.0),
-- Paid invoice from different month should not be included  
('invoice-004', 'INV-2025-004', 'client4@example.com', 8.0, 8.0, 0, 0, 'paid', '2025-05-15 10:00:00', '2025-05-15 10:00:00', '2025-05-15 10:00:00', 200000.0);